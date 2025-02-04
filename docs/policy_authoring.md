
# Policy authoring guide

This document is focused on the process of authoring policies. Together with the
[`examples` directory](../examples), it forms a tutorial that walks through policy
concepts as well how to test policies.

- [Policy authoring guide](#policy-authoring-guide)
  - [Policy syntax tutorial](#policy-syntax-tutorial)
    - [Simple policies part 1](#simple-policies-part-1)
    - [Simple policies part 2: Returning attributes](#simple-policies-part-2-returning-attributes)
    - [Advanced policies part 1](#advanced-policies-part-1)
    - [Advanced policies part 2: Adding compliant resource info](#advanced-policies-part-2-adding-compliant-resource-info)
    - [Advanced policies part 3: Correlating resources](#advanced-policies-part-3-correlating-resources)
    - [Advanced policies part 4: Correlating resources](#advanced-policies-part-4-correlating-resources)
    - [Advanced policies part 5: Returning attributes](#advanced-policies-part-5-returning-attributes)
    - [Missing resources](#missing-resources)
  - [Testing policies](#testing-policies)
    - [Creating and using test fixtures](#creating-and-using-test-fixtures)
    - [Using the REPL](#using-the-repl)
      - [With an input](#with-an-input)
        - [Examples](#examples)
      - [Without an input](#without-an-input)
        - [Example](#example)
    - [Using snapshot_testing.match](#using-snapshot_testingmatch)

## Policy syntax tutorial

We will walk through the policies in the examples directory, starting out with
simple policies and gradually adding concepts.

### Simple policies part 1

[examples/01-simple.rego](../examples/01-simple.rego)

### Simple policies part 2: Returning attributes

[examples/02-simple-attributes.rego](../examples/02-simple-attributes.rego)

### Advanced policies part 1

[examples/03-advanced.rego](../examples/03-advanced.rego)

### Advanced policies part 2: Adding compliant resource info

[examples/04-advanced-resources.rego](../examples/04-advanced-resources.rego)

### Advanced policies part 3: Correlating resources

[examples/05-advanced-primary-resource.rego](../examples/05-advanced-primary-resource.rego)

### Advanced policies part 4: Correlating resources

[examples/06-advanced-correlation.rego](../examples/06-advanced-correlation.rego)

### Advanced policies part 5: Returning attributes

[examples/07-advanced-attributes.rego](../examples/07-advanced-attributes.rego)

### Missing resources

[examples/08-missing.rego](../examples/08-missing.rego)

## Testing policies

### Creating and using test fixtures

In order to test policies, we want to generate _fixtures_ so that we freeze in
the processed input generated by the Unified Policy Engine.  This allows us
to use standard OPA tooling.

You can generate a fixture using the `fixture` command.  For example, we can
generate a fixture for the example terraform file we are using like this:

    ./policy-engine fixture examples/main.tf >examples/tests/fixture.json

Fixtures can also be generated using other applications.  The important bit is
that a fixture should provide a `mock_input` rule which represents the input to
be used for the test.

This allows us to import and use the fixture in a test:

[examples/tests/advanced-rule-test.rego](../examples/tests/advanced-rule-test.rego)

Running the tests:

    ./policy-engine -d examples test

We can also run using vanilla OPA.  This requires us to pass in the
[rego/](rego/) directory as well:

    opa test examples rego

### Using the REPL

Sometimes it's helpful to interactively evaluate policies in order to debug specific
portions of code. `policy-engine` includes a REPL that has two modes of
operation:

* With an input
* Without an input

Running with an input is intended to be used to debug policy code with some real input.
Running without an input is intended to be used to debug tests.

Both modes of operation use the ["pure rego" version](rego/snyk.rego) of the `snyk` API
rather than the custom built-ins used by the `run` command. In practice, these should
behave the same.

#### With an input

Running the REPL with an input will setup an environment that closely matches the way
rules are evaluated by:

* Parsing the input into a `State` object
* Setting the `input` document to the state object
  * This can be useful for inspecting the input from within the REPL, but policy code
    must use functions from the snyk API like snyk.resources() to access the input, to
    ensure compatibility with the production (non-repl) engine.

##### Examples

Introspecting a multi-resource policy:

```sh
# Invoking the REPL with an IaC input
$ ./policy-engine repl -d examples examples/main.tf
# Switching to the package of a multi-resource policy
> package rules.snyk_003.tf
# Evaluating the deny rule
> deny
[
  {
    "message": "Bucket names should not contain the word bucket, it's implied",
    "resource": {
      ...
    }
  },
  {
    "message": "Bucket names should not contain the word bucket, it's implied",
    "resource": {
      ...
    }
  }
]
# Evaluating parts of the policy. Both of these are defined in rules.snyk_003.tf
> has_bucket_name(buckets[0])
true
> 
```

Introspecting a single-resource policy:

```sh
# Invoking the REPL with an IaC input
$ ./policy-engine repl -d examples examples/main.tf
# Switching to the package of a single-resource policy
> package rules.snyk_001.tf
# Importing the snyk library so that we can use snyk.resources()
> import data.snyk
# Evaluating snyk.resources using the resource type defined in rules.snyk_001.tf
> snyk.resources(resource_type)
[
  {
    ...
    "_type": "aws_s3_bucket",
    ...
  },
  ...
]
# Evaluating the deny rule with a specific resource
> deny with input as snyk.resources(resource_type)[0]
[
  {
    "message": "Bucket names should not contain the word bucket, it's implied"
  }
]
> 
```

#### Without an input

Running the REPL without an input is useful for debugging tests and interacting with
test fixtures.

##### Example

```sh
# Invoking the REPL with a data directory that contains both policies and tests
$ ./policy-engine repl -d examples
# Switching to the package of a test. In this case, we're using the same package name
# for both the policy and the test in order to simplify the test code.
> package rules.snyk_003.tf
# Evaluating one of the tests
> test_policy
true
# Importing the fixture used in this test
> import data.examples.main
# Evaluating the deny for this policy with our test fixture
> deny with input as main.mock_input
[
  {
    "message": "Bucket names should not contain the word bucket, it's implied",
    "resource": {
      ...
    }
  },
  {
    "message": "Bucket names should not contain the word bucket, it's implied",
    "resource": {
      ...
    }
  }
]
> 
```

### Using snapshot_testing.match

Policy tests can be tedious to write and maintain.  We currently write the
expected output of the `deny` and `resources` rules, which can be quite large or
complex, by hand.  Any time we make updates, for example to reword the message
returned by a rule, we need to make the same update repeatedly in the expected
output in our tests.

This is where the `snapshot_testing.match` builtin comes in.  In your
`*_test.rego` files, you are encouraged to use the following style of tests:

```open-policy-agent
test_foo {
    some_variable = ...
    snyk.test.matches_snapshot(some_variable, "some/file/path.json")
}
```

This function will assert that the value of `some_variable` matches the contents
of the file `some/file/path.json` relative to the file that contains the
function call.

 *  If the file does not exist or the contents do not match, this function will
    return `false` and `policy-engine test` with print out the diff.
 *  If `policy-engine test` is run with the `--update-snapshots` option, this
    function will update any existing snapshots and create new ones.

The resulting snapshots should be checked in to version control
This saves a lot of time since now you only need to review the output of
policies, you don't need to manually write it down.
