package fargate

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jecoz/flexi"
)

const MaxBackoffMsec = 30000

type Fargate struct {
	sess   *session.Session
	client *ecs.ECS
}

func (f *Fargate) lazySession() *session.Session {
	if f.sess == nil {
		f.sess = session.Must(session.NewSession())
	}
	return f.sess
}

func (f *Fargate) lazyClient() *ecs.ECS {
	if f.client == nil {
		f.client = ecs.New(f.lazySession())
	}
	return f.client
}

func (f *Fargate) DescribeTask(ctx context.Context, cluster, arn string) (*ecs.Task, error) {
	input := &ecs.DescribeTasksInput{
		Cluster: stringPtr(cluster),
		Tasks:   stringPtrSlice([]string{arn}),
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	resp, err := f.lazyClient().DescribeTasksWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if len(resp.Tasks) == 0 {
		if len(resp.Failures) > 0 {
			return nil, fmt.Errorf("describe task: %v", resp.Failures[0].String())
		}
		return nil, fmt.Errorf("describe task: unable to fulfil request")
	}
	return resp.Tasks[0], nil
}

func (f *Fargate) waitRunningTask(ctx context.Context, cluster, arn string) (task *ecs.Task, err error) {
	// Stop when the context is invalidated or the response is no longer
	// successfull. We're waiting a pending process to become running here,
	// not to resume from a lost connection.
	for n := 1; ; n++ {
		// I borrowed this choice from
		// https://github.com/tailscale/tailscale/blob/abd79ea3685d41afbac5fb9d4c58c4423c60a409/logtail/backoff/backoff.go#L42
		msec := n * n * 10
		if msec > MaxBackoffMsec {
			msec = MaxBackoffMsec
		}
		wait := time.Duration(msec) * time.Millisecond
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			task, err = f.DescribeTask(ctx, cluster, arn)
			if err != nil {
				return
			}
			if *task.LastStatus == ecs.DesiredStatusRunning {
				return
			}
			// TODO: we could log each time we retry.
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			err = ctx.Err()
			return
		}

	}
}

type runTaskInput struct {
	Cluster        string
	TaskDefinition string
	Subnets        []string
	SecurityGroups []string
}

func (f *Fargate) RunTask(ctx context.Context, p runTaskInput) (*ecs.Task, error) {
	input := &ecs.RunTaskInput{
		Cluster:        stringPtr(p.Cluster),
		LaunchType:     stringPtr(ecs.LaunchTypeFargate),
		TaskDefinition: stringPtr(p.TaskDefinition),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				AssignPublicIp: stringPtr(ecs.AssignPublicIpEnabled),
				Subnets:        stringPtrSlice(p.Subnets),
				SecurityGroups: stringPtrSlice(p.SecurityGroups),
			},
		},
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	resp, err := f.lazyClient().RunTaskWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(resp.Tasks) == 0 {
		if len(resp.Failures) > 0 {
			return nil, fmt.Errorf("run task: %v", resp.Failures[0].String())
		}
		return nil, fmt.Errorf("run task: unable to fulfil request")
	}
	// TODO: what happens if resp contains more than one task? Is it possible?
	return resp.Tasks[0], nil
}

func (f *Fargate) StopTask(ctx context.Context, cluster, arn string) error {
	input := &ecs.StopTaskInput{
		Cluster: stringPtr(cluster),
		Task:    stringPtr(arn),
	}
	if err := input.Validate(); err != nil {
		return err
	}
	_, err := f.lazyClient().StopTaskWithContext(ctx, input)
	return err
}

func describeNetworkInterface(ctx context.Context, sess *session.Session, eni string) (*ec2.NetworkInterface, error) {
	// NOTE: this function uses EC2. If more functions like this are needed,
	// extract them into a separte ec2 package.
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: stringPtrSlice([]string{eni}),
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	resp, err := ec2.New(sess).DescribeNetworkInterfacesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(resp.NetworkInterfaces) == 0 {
		return nil, fmt.Errorf("no interface found for %v", eni)
	}
	return resp.NetworkInterfaces[0], nil
}

func eniFromTask(task *ecs.Task) (string, error) {
	if len(task.Attachments) == 0 {
		return "", fmt.Errorf("missing task attachments")
	}
	var eniAttach *ecs.Attachment
	for i, v := range task.Attachments {
		if *v.Type == "ElasticNetworkInterface" {
			eniAttach = task.Attachments[i]
			break
		}
	}
	if eniAttach == nil {
		return "", fmt.Errorf("missing ElasticNetworkInterface attachment")
	}
	var eni string
	for _, v := range eniAttach.Details {
		if *v.Name == "networkInterfaceId" {
			eni = *v.Value
			break
		}
	}
	if eni == "" {
		return "", fmt.Errorf("unable to find network interface id within eni attachment")
	}
	return eni, nil
}

func (f *Fargate) Spawn(ctx context.Context, t flexi.Task) (*flexi.RemoteProcess, error) {
	task, err := f.RunTask(ctx, runTaskInput{
		Cluster:        t.Image.Cluster,
		TaskDefinition: t.Image.Name,
		Subnets:        t.Image.Subnets,
		SecurityGroups: t.Image.SecurityGroups,
	})
	if err != nil {
		return nil, err
	}
	if task, err = f.waitRunningTask(ctx, t.Image.Cluster, *task.TaskArn); err != nil {
		return nil, err
	}

	eni, err := eniFromTask(task)
	if err != nil {
		return nil, err
	}
	ifi, err := describeNetworkInterface(ctx, f.lazySession(), eni)
	if err != nil {
		return nil, err
	}

	return &flexi.RemoteProcess{
		Addr:    net.JoinHostPort(*ifi.Association.PublicIp, t.Image.Service),
		Name:    *task.TaskArn,
		Cluster: t.Image.Cluster,
		Tags:    []string{"fargate", "9p"},
	}, nil
}

func (f *Fargate) Kill(ctx context.Context, p *flexi.RemoteProcess) error {
	return f.StopTask(ctx, p.Cluster, p.Name)
}

func stringPtrSlice(s []string) []*string {
	dst := make([]*string, len(s))
	for i := range s {
		dst[i] = stringPtr(s[i])
	}
	return dst
}

func stringPtr(s string) *string { return &s }
