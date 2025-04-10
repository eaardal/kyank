# Kyank

Yank things from Kubernetes.

Currently works with Pod environment variables.

## Motivation

Want to use some environment variables or even environment variables with secret values locally?
This tool finds them and lists them. By eval'ing the output, you can quickly set your local envs to mirror a pod's.

## Usage

Use `--help` to explore the options

```
NAME:
   kyank - Invoke with the Kubernetes namespace, Pod ID or Deployment name and at least one environment variable to read

USAGE:
   kyank [global options]

DESCRIPTION:
   Yank things from Kubernetes

GLOBAL OPTIONS:
   --namespace string, -n string                        Kubernetes namespace [$KYANK_K8S_NAMESPACE]
   --pod-id string, -p string                           Kubernetes Pod ID. Either Pod ID or Deployment name is required.
   --deployment string, -d string                       Kubernetes Deployment name. Either Deployment name or Pod ID is required.
   --env string, -e string [ --env string, -e string ]  Kubernetes pod environment variables
   --prefix string, -p string                           This text will be prepended to each environment variable line as output. Useful if you want to add 'export ' before each line. [$KYANK_PREFIX]
   --separator string, -s string                        The separator text between an environment variable's key and value text. By default '=' is used (KEY=VALUE), but if you want 'KEY: VALUE' or something else instead you can for example specify --separator ': ' (default: =) [$KYANK_SEPARATOR]
   --help, -h                                           show help
```

## Example

Given that we have the Pod `some-app-123-456` in Kubernetes with these environment variables:

```
ABC=aaa
DEF=bbb
FOO=foo
BAR=bar
```

Reading one environment variable by Pod ID:

```
> kyank --namespace apps --pod-id some-app-123-456 --env ABC
ABC=aaa
```

Shorthand:

```
> kyank -n apps -p some-app-123-456 -e ABC
ABC=aaa
```

Reading one environment variable by Deployment name:

```
> kyank --namespace apps --deployment some-app --env ABC
ABC=aaa
```

Reading many environment variables:

```
> kyank --namespace apps --pod-id some-app-123-456 --env ABC,DEF,FOO,BAR
ABC=aaa
DEF=bbb
FOO=foo
BAR=bar
```

The `--prefix` flag allows us to modify the output for scripting (note the trailing space after `export` included by purpose in the prefix text):

```
> kyank --namespace apps --pod-id some-app-123-456 --env ABC,DEF --prefix "export "
export ABC=aaa
export DEF=bbb
```

The output can then be eval'ed to set these as environment variables in your shell:

```
> eval "$(kyank --namespace apps --pod-id some-app-123-456 --env ABC --prefix "export ")"
> env | grep ABC
ABC=aaa
```

## Environment variables with secret values

Secret values are read as by default if needed. No need to do anything special.