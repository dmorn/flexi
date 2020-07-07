<!--
SPDX-FileCopyrightText: 2020 jecoz

SPDX-License-Identifier: BSD-3-Clause
-->

The `echo64spawner` and `echo64killer` packages use the `fargate` package to start `echo64` in an AWS ECS-managed container. Authentication and customization of how and where the AWS account operates is done through **environment variables**.