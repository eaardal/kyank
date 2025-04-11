package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
)

type options struct {
	k8sContextName string
	namespace      string
	podId          string
	deploymentName string
	envNames       []string
	prefix         string
	suffix         string
	separator      string
}

func main() {
	cmd := &cli.Command{
		Name:        "kyank",
		Description: "Yank things from Kubernetes",
		Usage:       "Invoke with the Kubernetes namespace, Pod ID or Deployment name and at least one environment variable to read",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Usage:   "Kubernetes context. This is optional, but helps ensure the command is being run against the exact correct kubernetes context",
				Aliases: []string{"c"},
				Sources: cli.EnvVars("KYANK_K8S_CONTEXT"),
			},
			&cli.StringFlag{
				Name:     "namespace",
				Usage:    "Kubernetes namespace",
				Aliases:  []string{"n"},
				Required: true,
				Sources:  cli.EnvVars("KYANK_K8S_NAMESPACE"),
			},
			&cli.StringFlag{
				Name:    "pod-id",
				Usage:   "Kubernetes Pod ID. Either Pod ID or Deployment name is required.",
				Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:    "deployment",
				Usage:   "Kubernetes Deployment name. Either Deployment name or Pod ID is required.",
				Aliases: []string{"d"},
			},
			&cli.StringSliceFlag{
				Name:     "env",
				Usage:    "Kubernetes pod environment variables",
				Aliases:  []string{"e"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "prefix",
				Usage:   "This text will be prepended to each environment variable line as output. Useful if you want to add 'export ' before each line.",
				Sources: cli.EnvVars("KYANK_PREFIX"),
			},
			&cli.StringFlag{
				Name:    "suffix",
				Usage:   "This text will be appended to each environment variable line as output.",
				Sources: cli.EnvVars("KYANK_SUFFIX"),
			},
			&cli.StringFlag{
				Name:        "separator",
				Usage:       "The separator text between an environment variable's key and value text. By default '=' is used (KEY=VALUE), but if you want 'KEY: VALUE' or something else instead you can for example specify --separator ': '",
				Aliases:     []string{"s"},
				DefaultText: "=",
				Value:       "=",
				Sources:     cli.EnvVars("KYANK_SEPARATOR"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			opts := options{
				k8sContextName: c.String("context"),
				namespace:      c.String("namespace"),
				podId:          c.String("pod-id"),
				deploymentName: c.String("deployment"),
				envNames:       c.StringSlice("env"),
				prefix:         c.String("prefix"),
				suffix:         c.String("suffix"),
				separator:      c.String("separator"),
			}

			if opts.podId == "" && opts.deploymentName == "" {
				return fmt.Errorf("you must specify a pod id or deployment")
			}

			api := newK8sApi(opts.k8sContextName, opts.namespace)
			if err := api.init(); err != nil {
				return err
			}

			if opts.podId != "" {
				lines, err := lookupPodEnvironmentVariables(ctx, opts, api)
				if err != nil {
					return err
				}

				printStdout(stringifyLines(lines))
				return nil
			}

			if opts.deploymentName != "" {
				lines, err := lookupDeploymentEnvironmentVariables(ctx, opts, api)
				if err != nil {
					return err
				}

				printStdout(stringifyLines(lines))
				return nil
			}

			return fmt.Errorf("unable to lookup environment variables with the provided arguments")
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		printStderr("kyank crashed: %v", err)
	}
}

func lookupPodEnvironmentVariables(ctx context.Context, opts options, api *k8sApi) ([]string, error) {
	pod, err := api.getPod(ctx, opts.podId)
	if err != nil {
		return nil, err
	}

	return extractAndFormatContainerEnvironmentVariables(ctx, pod.Spec.Containers, opts, api)
}

func lookupDeploymentEnvironmentVariables(ctx context.Context, opts options, api *k8sApi) ([]string, error) {
	depl, err := api.getDeployment(ctx, opts.deploymentName)
	if err != nil {
		return nil, err
	}

	return extractAndFormatContainerEnvironmentVariables(ctx, depl.Spec.Template.Spec.Containers, opts, api)
}

func extractAndFormatContainerEnvironmentVariables(ctx context.Context, containers []v1.Container, opts options, api *k8sApi) ([]string, error) {
	var lines []string

	for _, container := range containers {
		envs, err := matchEnvironmentVariables(ctx, container.Env, opts.envNames, api)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve environment variables: %v", err)
		}

		for key, value := range envs {
			keyValue := fmt.Sprintf("%s%s%s", key, opts.separator, value)
			line := prependString(opts.prefix, keyValue)
			line = appendString(opts.suffix, line)
			lines = append(lines, line)
		}
	}

	return lines, nil
}

func matchEnvironmentVariables(ctx context.Context, envVars []v1.EnvVar, envsToLookFor []string, api *k8sApi) (map[string]string, error) {
	envs := make(map[string]string)

	for _, envVar := range envVars {
		for _, envToLookup := range envsToLookFor {
			if envVar.Name == envToLookup {
				value, err := resolvePlainTextOrSecretEnvValue(ctx, envVar, api)
				if err != nil {
					return nil, err
				}
				envs[envToLookup] = value
			}
		}
	}

	return envs, nil
}

func resolvePlainTextOrSecretEnvValue(ctx context.Context, podEnv v1.EnvVar, api *k8sApi) (string, error) {
	if podEnv.Value != "" {
		return podEnv.Value, nil
	}

	if podEnv.ValueFrom != nil && podEnv.ValueFrom.SecretKeyRef != nil {
		secretName := podEnv.ValueFrom.SecretKeyRef.Name
		secretKey := podEnv.ValueFrom.SecretKeyRef.Key
		return api.getSecretValue(ctx, secretName, secretKey)
	}

	return "", fmt.Errorf("unable to resolve environment variable value for %s", podEnv.Name)
}

func stringifyLines(lines []string) string {
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
	}
	return sb.String()
}

func prependString(prefix string, value string) string {
	if prefix != "" {
		return prefix + value
	}
	return value
}

func appendString(suffix string, value string) string {
	if suffix != "" {
		return value + suffix
	}
	return value
}

func printStdout(format string, a ...any) {
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(format, a...))
}

func printStderr(format string, a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
}
