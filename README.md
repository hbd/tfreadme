# TF Module README Generator

## Overview

Generate READMEs for Terraform modules from tf files.

## How To Install

`go get -u github.com/hbd/tfreadme`

or

`git clone git@github.com:hbd/tfreadme.git && make install`

## How To Use

### v0.0.1

`tfreadme` expects

* input vars from `variables.tf`

* outputs from `outputs.tf`

from within the current directory. The README content is output to stdout.

`cd` into the tf module directory and run
`tfreadme > README.md`

## Example README

``` markdown

# IAM Terraform Module

## Overview


## Input

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| my_var | My awesome var. | string |  |  no  |

## Output

| Name | Description | Type |
|------|-------------|:----:|
| my_output | My awesome output. |  |

## Usage

```

```

## Troubleshooting


```
