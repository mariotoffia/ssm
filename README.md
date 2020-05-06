# Introduction
This library is intended to allow for encode / decode _go_ `struct` _fields_ from [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) and [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).

This library do not yet even have a 0.0.1 release and hence in non usable state. It basically now can do a plain `Unmarshal` & `Marshal` operation, with PMS and ASM, partially or fully with reporting of which fields did not have any PMS counterpart. It also supports Filtering for selective unmarshal / marshal _pms_ and _asm_ fields.

Only string value (**not binary**) for Secrets Manager is currently supported!

The intention to this library to simplify fetching & upserting one or more parameters, secrets blended with other settings. It is also intended to be as efficient as possible and hence possible to filter, exclude or include, which properties that should participate in `Unmarshal` or `Marshal` operation. It uses go standard _Tag_ support to direct the `Serializer` how to `Marshal` or `Unmarshal` the data. For example

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout"`
  Db struct {
    ConnectString string `asm:"connection, prefix=global/accountingdb"`
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
+ /eap/global/accountingdb/connection (Secrets Manager)
+ /eap/test-service/timeout (Parameter Store)
+ /eap/test-service/db/batchsize (Parameter Store)
+ /eap/test-service/db/timeout (Parameter Store)

# Standard Usage

## Good For Lambda Configuration
In combination with [env](https://github.com/codingconcepts/env) this is a great way of centrally adminitrating your configuration but allow override of those using environment variables. For example
```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, keyid=default, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
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

Note that plain `Unmarshal` will examine the structs for **both** _asm_ and _pms_ tags. If you want to control, and optimize speed and remote manager access, use `UnmarshalWithOpts` whey you may specify which tag types to use in the unmarshal operation.

Since the `keyid=default` is specifies (if a write operation and key do not exists) that the account default CMK is used.

## AWS Secrets Manager
In addition to Systems Manager, Parameter Store, this serializer can handle _asm_ tags that referes to the Secrets Manager instead. This is good if you e.g. have a shared secret for a RDS and wish to rotate the secret. For example, if we would use PMS for all configuration around how to handle the database and logic around it and then use the secrets manager for the actual connection string. It could look like this:

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `asm:"connection, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
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

Just a simple _pms_ to _asm_ tag substitution and now the connection string is managed in the secrets manager. Since `Unmarshal`, by default, unmarshals both _asm_ and _psm_ no changes in the unmarshal code is needed. 

You may if you wish only access the secret or parameters using the unmarshal directives `OnlyPsm` or `OnlyAsm` in the `UnmarshalWithOpts` method. For example

```go
if _, err := s.UnmarshalWithOpts(&ctx, NoFilter, OnlyPms); err != nil  {
  panic()
}
```

The above will only unmarshal the parameter store data (by specifying `OnlyPms`) and **not** secrets manager `ConnectionString`. Hence it would be _""_. This can of course be achieved by _Filters_ (see below) but is a tiny bit optimization if you know that an entire remote store is not needed. In contrast, using filters you may selectively unmarshal values from both _asm_ and _pms_ (see filter below).

### Versions
Since AWS Secrets Manager handles versions in two ways and they are mutual exclusive, you only may specify one of the following _vs_, see [Version Stage](https://docs.aws.amazon.com/secretsmanager/latest/userguide/terms-concepts.html#term_staging-label), and _vid_ (Version Id). If none is specified the _AWSCURRENT_ staging label is used as _vs_ and hence the last version is retrieved.

```go
type AlwaysLatest struct {
  ConnectString string `asm:connection, vs=AWSCURRENT"`
}
```
The above example explicit states that this property will always be attached to lastest version since the Version Stage is always point to _AWSCURRENT_ stage label.

```go
type AlwaysLatest struct {
  ConnectString string `asm:connection, vs=AWSCURRENT"`
}

type AlwaysPrevious struct {
  ConnectString string `asm:connection, vs=PREVIOUS"`
}

// Set and Marshal AlwaysLatest
// Set and Marshal AlwaysLatest
// Unmarshal AlwaysPrevious - will contain the previous value in ConnectString
// Unmarshal AlwaysLatest - will contain the current value in ConnectString
```


## Filters
If you don't want all properties to be set (faster response-times) use a filter to include & exclude properties. Filters also work in the hiarchy, i.e. you may set a exclusion for on a field that do have nested sub-structs beneach
and all of those will be automatically excluded. However, you may override that both on tree level or explicit on leaf (a specific field property that is *not* a sub-struct). For example

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, keyid=default, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Flow          struct {
      Base  int `pms:"base"`
      Prime int `pms:"prime"`
    }
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
if _, err := s.UnmarshalWithOpts(&ctx,
              support.NewFilters().
                      Exclude("Db").
                      Include("Db.ConnectString").
                      Include("Db.Flow"), OnlyPms); err != nil  {
  panic()
}

fmt.Printf("got total timeout of %d and connect using %s (base: %d, prime %d)", 
    ctx.TotalTimeout, ctx.Db.ConnectString, ctx.Db.Flow.Base, ctx.Db.Flow.Prime)

fmt.Printf("No data for BatchSize %d and DbTimeout %d", ctx.Db.BatchSize, ctx.Db.DbTimeout)
```
The above sample will first _Exclude_ everything beneath the Db node. But since we have explicit (Leaf) and Node implicit Includes *beneath* the exclusion, those properties will be included. In this case `ConnectString`, everything beneatch `Flow` is included. However, everything else beneath `Db` is excluded, including `BatchSize` and `DbTimeout`.

It also used the `OnlyPms` to illustrate that you may select what types of tags the unmarshaller shall use. In this case it is only a very scarse bit of optimization. However, if you would have _asm_ tags in this struct it would access the Secrets Manager if not filtered out.

## Taking Care of Not Backed Parameters
If there was no backing parameter on e.g. Parameter Store, the `Unmarshal` methods will return as `map[string]support.FullNameField`. The map is keyed with the field navigation e.g. _Db.Flow.Base_ would refer to the
```go
Base  int `pms:"base"`
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
    ConnectString string `pms:"connection, keyid=default, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Missing       string `pms:"missing-backing-field"`
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

# Writing (Marshalling)
It is possible to marshal using the structs towards the Parameter Store and Secrets Manager. To be smart and not update all parameters / secrets use filter to include and exclude struct fields to be marshalled. Note that writing to secrets manager and read back the values directly may return some secrets with the old values since it seems that it uses eventual consistency and hence a later point in time you get the new values.

Marshalling is quite simple, just pass the pointer to the struct that you wish to marshal and it will iterate the fields and any sub-structs. The **error** mechanism is a little bit different. It will always only return the `support.FullNameField` (zero or more). If any geric error a **single** `support.FullNameField` is returned with the _Error_ property set to the error encountered. It has no _LocalName_ etc. set, just the _Error_ field. When it fails for some reason to write a field, it is returned as with `Unmarshal`. However, the _Error_ field is always set to the last exception encountered. This exception may not be the source since retries.

Marshal is really a _Upsert_ operation where it tries to create, if already existant it will _Update_ the parameter. If update; it will then check if any tags are associated with the field and set those tags on the parameter / secret.

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, keyid=default, prefix=global/accountingdb"`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Missing       string `pms:"missing-backing-field"`
  }
}

ctx := MyContext { Caller: "kalle", 
// initalize the struct and substruct ...
}

s := ssm.NewSsmSerializer("eap", "test-service")
err := s.Marshal(&ctx)
if len(err) > 0
  panic()
}
```

This above example marshals the **entire** struct and it's sub-struct field. Since Parameter Store do not have a batch mechanism the parameters are created / updated one by one. Hence this `Marshal` operation will call the parameter store 5 times (since no tags are present). Since the `ConnectString` is decorated with a _keyid_ it will be encrypted (in this case using the account default KMS key for parameter store).

Therefore make sure to use _filter_ to narrow down the parameter that you really wanted to write!

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, keyid=default, prefix=global/accountingdb"`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Flow          struct {
      Base  int `pms:"base"`
      Prime int `pms:"prime"`
    }
  }
}

ctx := MyContext { Caller: "kalle", 
// initalize the struct and substruct ...
}

s := ssm.NewSsmSerializer("eap", "test-service")
err := s.MarshalWithOpts(&ctx,
          support.NewFilters().
              Exclude("Db").
              Include("Db.ConnectString").
              Include("Db.Flow"), OnlyPms); err != nil  {

if len(err) > 0 {
  panic()
}
```
The above example will only creater / update the parameter store with three parameters
+ ConnectString
+ Base
+ Prime

Hence, only 3 invocations is done for this operation instead of six.

It is also, as with `Unmarshal` blend _asm_ and _pms_ tags and the serializer will marshal towards parameter store or secrets manager respectively.

```go
type MyContext struct {
  Caller        string
  TotalTimeout  int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `asm:"connection, keyid=default, prefix=global/accountingdb"`
    BatchSize     int `pms:"batchsize"`
    DbTimeout     int `pms:"timeout"`
    UpdateRevenue bool
    Signer        string
    Missing       string `pms:"missing-backing-field"`
  }
}

ctx := MyContext { Caller: "kalle", 
// initalize the struct and substruct ...
}

s := ssm.NewSsmSerializer("eap", "test-service")
err := s.Marshal(&ctx)
if len(err) > 0
  panic()
}
```

Again, this will **bluntly** _Marshal_ all parameters in struct. Since _ConnectionString_ is in the Secrets Manager it could possibly incur three invocations. If not already existant it will only use one create. But if already existant it tries to create, if fails it will update. If tags are present it will also invoke a tag resource. In above example, since missing tags, it will use one or two invocations. 

_ - It's better to use filters :)_
