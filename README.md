# Kyank

Yank things from Kubernetes.

Currently works with Pod environment variables.

## Motivation

Want to use some environment variables or even environment variables with secret values locally?
This tool finds them and lists them. By eval'ing the output, you can quickly set your local envs to mirror a pod's.

## Example

Given that we have the Pod `some-app-123-456` in Kubernetes with these environment variables:

```
ABC=aaa
DEF=bbb
FOO=foo
BAR=bar
```

Specifying one environment variable:

```
> kyank --namespace apps --pod-id some-app-123-456 --env ABC
ABC=aaa
```

Specifying many environment variables:

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