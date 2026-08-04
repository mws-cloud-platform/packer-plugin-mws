// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
)

type subSteps []subStep

func (s subSteps) run(ctx context.Context, ui packer.Ui) error {
	for _, step := range s {
		if !step.check() {
			continue
		}
		ui.Sayf("Creating %s...", step)
		err := step.run(ctx)
		if err != nil {
			return fmt.Errorf("create %s: %w", step, err)
		}
		ui.Sayf("%s created", common.UpperFirst(step.String()))
	}

	return nil
}

type subStep interface {
	check() bool
	run(ctx context.Context) error
	String() string
}

type subStepWithoutResult[Params any] struct {
	cond         bool
	resourceType string
	resourceName string
	action       func(context.Context, Params) error
	params       Params
}

func (s *subStepWithoutResult[Params]) check() bool {
	return s.cond
}

func (s *subStepWithoutResult[Params]) run(ctx context.Context) error {
	return s.action(ctx, s.params)
}

func (s *subStepWithoutResult[Params]) String() string {
	return fmt.Sprintf("%s %q", s.resourceType, s.resourceName)
}

type subStepWithResult[Res, Params any] struct {
	cond         bool
	resourceType string
	resourceName string
	action       func(context.Context, Params) (*Res, error)
	result       *Res
	params       Params
}

func (s *subStepWithResult[Res, Params]) check() bool {
	return s.cond
}

func (s *subStepWithResult[Res, Params]) run(ctx context.Context) error {
	result, err := s.action(ctx, s.params)
	if err != nil {
		return err
	}
	*(s.result) = *result
	return nil
}

func (s *subStepWithResult[Res, Params]) String() string {
	return fmt.Sprintf("%s %q", s.resourceType, s.resourceName)
}
