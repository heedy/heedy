# Tutorial

Heedy's timeseries transform capabilities are simple, but powerful.
This tutorial will guide you from the very basics of transforming data to creating sophisticated transform pipelines.

## Basics

Each datapoint in a timeseries has the following format:

```
{
    "t": floating point timestamp (unix time in seconds),
    "dt": floating point duration in seconds,
    "d": the datapoint's data content.
}
```

Unless explicitly stated, transforms focus on the datapoint's data content. To start out, we will use the following tiny timeseries (which has no duration):

```json
[
  { "t": 123, "d": 2 },
  { "t": 124, "d": 3 },
  { "t": 125, "d": 0.1 },
  { "t": 126, "d": -50 }
]
```

### Comparisons

Let's check which datapoints have their data >= 1.

```
$ >= 1
```

If you are familiar with programming, this is just a simple comparison statement. In PipeScript, \$ represents the "current datapoint". The transform is then run on each consecutive datapoint in the timeseries:

```json
[
  { "t": 123, "d": true },
  { "t": 124, "d": true },
  { "t": 125, "d": false },
  { "t": 126, "d": false }
]
```

You can also use `and`, `or`, and `not` to create logic of arbitrary complexity:

```
$ < 0 or not $ < 1
```

```json
[
  { "t": 123, "d": true },
  { "t": 124, "d": true },
  { "t": 125, "d": false },
  { "t": 126, "d": true }
]
```

### Algebra

PipeScript also supports basic algebra. In particular, `+-/*%^` are all built into the language, with `x^y` meaning `pow(x,y)`.

```
($+5)/2
```

gives:

```json
[
  { "t": 123, "d": 3.5 },
  { "t": 124, "d": 4 },
  { "t": 125, "d": 2.55 },
  { "t": 126, "d": -22.5 }
]
```

### Aggregating Data

Not all transforms return an answer for each datapoint. Some _aggregate_ your data:

```
sum
```

```json
[{ "t": 123, "dt": 3, "d": -44.9 }]
```

As expected, the sum transform returns a single datapoint, which has in its data portion the sum over the entire timeseries. Examples of other available aggregators are `count`, `mean`, `min`, and `max`.

### Filtering Data

Transforms can take arguments as input. For example, the `filter` transform removes all datapoints that don't satisfy the condition given in its first argument:

```
filter($>=2)
```

```json
[
  { "t": 123, "d": 2 },
  { "t": 124, "d": 3 }
]
```

Note that the parentheses are optional here. The above transform is equivalent to:

```
filter $>=2
```

## Chaining Transforms

Oftentimes, you want to combine multiple transforms. PipeScript allows you to do this using pipes:

```
a | b | c | d
```

The output of `a` is used as the input to `b`, and so forth.

For this section, we will use the following dataset from a fitness tracker:

```json
[
  {
    "t": 1,
    "d": {
      "steps": 14,
      "activity": "walking"
    }
  },
  {
    "t": 2,
    "d": {
      "steps": 10,
      "activity": "running"
    }
  },
  {
    "t": 3,
    "d": {
      "steps": 12,
      "activity": "walking"
    }
  },
  {
    "t": 4,
    "d": {
      "steps": 5,
      "activity": "running"
    }
  }
]
```

Suppose we want to get the **total number of steps we took while running**. We cannot do this with one transform, but by chaining together a couple simple transforms, we can get there!

First off, let's filter the datapoints so that we have just those where we were running:

```
filter $("activity")=="running"
```

Notice that the \$ accepts an argument - it allows you to return a sub-object of the datapoint. Our result is:

```json
[
  {
    "t": 2,
    "d": {
      "steps": 10,
      "activity": "running"
    }
  },
  {
    "t": 4,
    "d": {
      "steps": 5,
      "activity": "running"
    }
  }
]
```

We can now add a `|` after the first part of our statement, and we can perform further transforms on the result of the previous operation (shown above).
After extracting only the datapoints that have their activity as "running", we return only the "steps" portion of the datapoint:

```
filter $("activity")=="running" | $("steps")
```

```json
[
  {
    "t": 2,
    "d": 10
  },
  {
    "t": 4,
    "d": 5
  }
]
```

Finally, we want to sum the datapoints to get the total number of steps while running:

```
filter $("activity")=="running" | $("steps") | sum
```

```json
[
  {
    "t": 2,
    "dt": 2,
    "d": 15
  }
]
```

## Advanced Pipes

All arguments to each transform are actually transform pipelines. For example, one can go multiple levels into a nested object within an argument to the filter transform:

```
filter( ($("level1") | $("level2")) == 4 )
```

For convenience, PipeScript also includes `:` as a pipe symbol with high prescedence (the pipe will be taken before algebra is done) which can allow you to simplify your script a bit by dropping the internal parentheses:

```
filter( $("level1"):$("level2") == 4 )
```

In order for the parent (`filter`) to always get SOME result in its argument, sub-transforms cannot include transforms that are not One-To-One (for each datapoint that they get as input, they output one datapoint). This means that you cannot nest `filter` transforms.

### Pipe Args

Some advanced transforms have an argument listed as a `pipe` type. Sub-transforms are linked to previous data, whereas pipes are treated as scripts, which are managed by the transform.

As an example, the `map` transform splits the timeseries along unique values of its first argument, and runs the pipe given in its second argument on each resulting stream.

We will once again use the timeseries from the previous example:

```json
[
  {
    "t": 1,
    "d": {
      "steps": 14,
      "activity": "walking"
    }
  },
  {
    "t": 2,
    "d": {
      "steps": 10,
      "activity": "running"
    }
  },
  {
    "t": 3,
    "d": {
      "steps": 12,
      "activity": "walking"
    }
  },
  {
    "t": 4,
    "d": {
      "steps": 5,
      "activity": "running"
    }
  }
]
```

Remember that previously, we found the total number of steps while running with the transform `filter $("activity")=="running" | $("steps") | sum`.

We will now extend that to find the number of steps for each activity, using the `map` transform:

```
map( $("activity") , $("steps"):sum )
```

```json
[
  {
    "t": 4,
    "dt": 3,
    "d": {
      "walking": 26,
      "running": 15
    }
  }
]
```

**What happened here?**

When calling `map(arg1,arg2)`, the `map` transform uses `arg2` as a Pipe. It then makes copies of `arg2` for each value of `arg1`, returning the output of the pipe for each instantiation/

To clarify, we will see exactly what happened in the above call:

1. The `map` transform saw the first datapoint. The value of `arg1`, (`$("activity")`) was `walking`. It created a new instance of `arg2`, `$("steps"):sum`, and sent the datapoint through this transform, giving a total of `14` so far for `walking`.
2. The next datapoint had as its activity `running`. Another new instance of `$("steps"):sum` was created, and the datapoint was sent through it. The sum for `running` starts at `10`
3. The third datapoint is `walking`. `map` already has a pipeline started for this value, so it passes the new datapoint through the first pipe, giving a sum of `26` (14+12)
4. The fourth datapoint is `running`. Passing it to the the corresponding pipe, we get `15`.
5. There are no more datapoints. The `map` transform returns an object with the last value of each pipe as a result.

The ability of transforms to be passed pipes as arguments allows them to do extremely powerful aggregations.

## Object-Valued Transforms

You might want to perform multiple calculations at once in pipescript, or perhaps simply return an object. For this reason, PipeScript supports JSON-like values:

```
{"sum": sum, "total": count}
```

This transform will return both the sum of all of the datapoints' values, and the number of datapoints at the same time. This object support also enables you to save values for later use in the pipeline.

Finally, since transforms can get fairly complex with objects, PipeScript does accept multiline scripts. That is, the following is a valid script format:

```
filter $("activity")!="still"
| {
    "total": $("steps"):sum,
    "some_random_stuff": ( $("steps") | something | something else )
}
```
