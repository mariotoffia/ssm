# Introduction
This library is intended to allow for encode / decode _go_ `struct` _fields_ from [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) and [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).

This library do not yet even have a 0.0.1 release and hence in non usable state. It basically now can do a plain `Unmarshal` operation, with PMS, partially or fully with reporting of which fields did not have any PMS counterpart.

The intention to this library to simplify fetching one or more parameters, secrets blended with other settings. It is also intended to be as efficient as possible and hence possible to filter, exclude or include, which properties that should participate in `Unmarshal` or `Marshal` operation. It uses go standard _Tag_ support to direct the `Serializer` how to `Marshal` or `Unmarshal` the data. For example

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb"`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
_, err := s.Unmarshal(&ctx)
if err != nil {
  panic()
}

fmt.Printf("got total timeout of %d and connect using %s ...", ctx.TotalTimeout, ctx.Db.ConnectString)
```

The above example shows how to blend _PMS_ backed data with data set by the service itself to perform the work. Note that the `ConnectString` is a global setting and hence independant on the service it will be retrieved from _/{env}/global/accountingdb/connection_ parameter. In this way it is possible to constrain parameters to a single service, share between services or have notion of global parameters. Environment is *always* present, thus mandatory.

The above example uses keys from 
+ /eap/global/accountingdb/connection
+ /eap/test-service/timeout
+ /eap/test-service/db/batchsize
+ /eap/test-service/db/timeout

# Standard Usage

## Good For Lambda Configuration
In combination with [env](https://github.com/codingconcepts/env) this is a great way of centrally adminitrating your configuration but allow override of those using environment variables. For example
```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
if _, err := s.Unmarshal(&ctx); err != nil  {
  panic()
}

if err := env.set(&ctx); err != nil  {
  panic()
}
// If we e.g. set the TOTAL_TIMEOUT = 99 in the env for the lambda 
// the ctx.TotalTimeout will be 99 and hence overridden locally
fmt.Printf("got total timeout of %d and connect using %s ...", ctx.TotalTimeout, ctx.Db.ConnectString)
```

## Filters
If you don't want all properties to be set (faster response-times) use a filter to include & exclude properties. Filters also work in the hiarchy, i.e. you may set a exclusion for on a field that do have nested sub-structs beneach
and all of those will be automatically excluded. However, you may override that both on tree level or explicit on leaf (a specific field property that is *not* a sub-struct). For example

```go

type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Flow          struct {
      Base  int `pms:base`
      Prime int `pms:prime`
    }
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
if _, err := s.UnmarshalFilterd(&ctx,
              support.NewFilters().
                      Exclude("Db").
                      Include("Db.ConnectString").
                      Include("Db.Misc")); err != nil  {
  panic()
}

fmt.Printf("got total timeout of %d and connect using %s (base: %d, prime %d)", 
    ctx.TotalTimeout, ctx.Db.ConnectString, ctx.Db.Flow.Base, ctx.Db.Flow.Prime)

fmt.Printf("No data for BatchSize %d and DbTimeout %d", ctx.Db.BatchSize, ctx.Db.DbTimeout)
```
The above sample will first _Exclude_ everything beneath the Db node. But since we have explicit (Leaf) and Node implicit Includes *beneath* the exclusion, those properties will be included. In this case `ConnectString`, everything beneatch `Flow` is included. However, everything else beneath `Db` is excluded, including `BatchSize` and `DbTimeout`.

# Enhanced Usage

## Taking Care of Not Backed Parameters
If there was no backing parameter on e.g. Parameter Store, the `Unmarshal` methods will return as `map[string]support.FullNameField`. The map is keyed with the field navigation e.g. _Db.Flow.Base_ would refer to the
```go
Base  int `pms:base`
```
parameter under the `Flow` field. The value is a `FullNameField struct` where it contains
```go
type FullNameField struct {
	// Local name in dotted navigation format
	LocalName string
	// Remove name as required by AWS
	// (for PMS this is not a ARN)
	RemoteName string
	// The field within the struct that is reffered
	Field reflect.StructField
	// The value accessor to the field. Note if this is
	// a pointer; it may not have a value do check IsValid
	// before accessing.
	Value reflect.Value
}
```

The `LocalName` is the same as the `map` key. `RemoteName` is the _AWS_ specific remote name. In parameter store it may e.g. be _/eap/test-service/db/flow/base_. In order to examine which field it is the `reflect.StructField` is included along with the `reflect.Value` to provide set and get accessors to the value itself.

This may be used to otherwise get or perform some default configuration for the reported fields. For example you may use a backing _JSON_ file within the service to read-up some sensible defaults and set those.
(I'll implement a default option for this so that the library may resort to do such and just report missing and what compensating (read backing JSON) action it took).

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Missing       sing `pms:"missing-backing-field"`
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
invalid, err := s.Unmarshal(&ctx)
if err != nil
  panic()
}

for key, fld := range invalid {
  fmt.Printf("Missing %s RemoteName %s\n", key, fld.RemoteName)
}

```

The above example will output ```Missing Db.Missing RemoteName /eap/test-service/db/missing-backing-field```.