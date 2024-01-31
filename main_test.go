package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/ONSdigital/log.go/v2/log"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-upload-service/features/steps"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")
var loggingFlag = flag.Bool("logging", false, "print logging")

type ComponentTest struct {
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	if !*loggingFlag {
		buf := bytes.NewBufferString("")
		log.SetDestination(buf, buf)
	}

	log.Namespace = "dp-upload-service"

	component := steps.NewUploadComponent()

	apiFeature := componenttest.NewAPIFeature(component.Initialiser)
	component.ApiFeature = apiFeature

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		component.Reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, e error) (context.Context, error) {
		err := component.Close()
		if err != nil {
			log.Error(ctx, "error closing service", err)
		}
		return ctx, nil
	})

	apiFeature.RegisterSteps(ctx)
	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		fmt.Println("=================================")
		fmt.Printf("Component test coverage: %.2f%%\n", testing.Coverage()*100)
		fmt.Println("=================================")

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
