// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

package fargate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/jecoz/flexi"
	"github.com/jecoz/flexi/file"
)

const LastStatusPollInterval = time.Millisecond * time.Duration(500)

type Fargate struct {
	// BackupDir is path pointing to the disk location where
	// Fargate will store the information about the spawned
	// tasks. In case of a recovery, files can be retriven
	// using LS.
	BackupDir string
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
	for {
		timer := time.NewTimer(LastStatusPollInterval)
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

type RunTaskInput struct {
	Cluster        string
	TaskDefinition string
	Subnets        []string
	SecurityGroups []string
}

func (f *Fargate) RunTask(ctx context.Context, p RunTaskInput) (*ecs.Task, error) {
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

func (f *Fargate) Spawn(ctx context.Context, r io.Reader) (*flexi.RemoteProcess, error) {
	var t Task
	if err := json.NewDecoder(r).Decode(&t); err != nil {
		return nil, fmt.Errorf("decoding task: %w", err)
	}
	task, err := f.RunTask(ctx, RunTaskInput{
		Cluster:        t.Image.Cluster,
		TaskDefinition: t.Image.Name,
		Subnets:        t.Image.Subnets,
		SecurityGroups: t.Image.SecurityGroups,
	})
	if err != nil {
		return nil, err
	}

	// If an error occours from this point on, we need to
	// stop the task too.
	undo := true
	defer func() {
		if !undo { return }
		// Even though the original context was invalidated, we need to
		// ensure we're not leaking resources.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		f.StopTask(ctx, t.Image.Cluster, *task.TaskArn)
	}()

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

	addr := net.JoinHostPort(*ifi.Association.PublicIp, t.Image.Service)
	name := *task.TaskArn

	bk, err := f.OpenBackup(name)
	if err != nil {
		return nil, fmt.Errorf("open backup: %w", err)
	}
	defer bk.Close()

	var b bytes.Buffer
	w := io.MultiWriter(bk, &b)

	if err := json.NewEncoder(w).Encode(&Container{
		Addr:    addr,
		Name:    name,
		Cluster: t.Image.Cluster,
	}); err != nil {
		return nil, err
	}

	undo = false
	return &flexi.RemoteProcess{Addr: addr, Name: name, Spawned: &b}, nil
}

func (f *Fargate) OpenBackup(arn string) (io.ReadWriteCloser, error) {
	return os.Create(filepath.Join(f.BackupDir, arn))
}

func (f *Fargate) RemoveBackup(arn string) error {
	return os.RemoveAll(filepath.Join(f.BackupDir, arn))
}

func (f *Fargate) Kill(ctx context.Context, r io.Reader) error {
	var p Container
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return err
	}
	if err := f.StopTask(ctx, p.Cluster, p.Name); err != nil {
		return err
	}
	if err :=  f.RemoveBackup(p.Name); err != nil {
		return fmt.Errorf("remove backup: %w", err)
	}
	return nil
}

func (f *Fargate) LS() ([]*flexi.RemoteProcess, error) {
	files := file.DiskLS(f.BackupDir)()
	rp := make([]*flexi.RemoteProcess, 0, len(files))
	for i, v := range files {
		rwc, err := v.Open()
		if err != nil {
			return nil, fmt.Errorf("LS file %d: open error: %w", i, err)
		}
		defer rwc.Close()

		var container Container
		var b bytes.Buffer
		tee := io.TeeReader(&b, rwc)
		if err := json.NewDecoder(tee).Decode(&container); err != nil {
			return nil, fmt.Errorf("LS file %d: error unmarshaling file content: %w", err)
		}
		rp = append(rp, &flexi.RemoteProcess{
			Name: container.Name,
			Addr: container.Addr,
			Spawned: &b,
		})
	}
	return rp, nil
}

func stringPtrSlice(s []string) []*string {
	dst := make([]*string, len(s))
	for i := range s {
		dst[i] = stringPtr(s[i])
	}
	return dst
}

func stringPtr(s string) *string { return &s }
