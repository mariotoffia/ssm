# Introduction
This library is intended to allow for encode / decode _go_ `struct` _fields_ from [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) and [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).

This library do not yet even have a 0.0.1 release and hence in non usable state. It basically now can do a plain `Unmarshal` operation, with PMS, partially or fully with reporting of which fields did not have any PMS counterpart.

The intention to this library to simplify fetching one or more parameters, secrets blended with other settings. It is also intended to be as efficient as possible and hence possible to filter, exclude or include, which properties that should participate in `Unmarshal` or `Marshal` operation. It uses go standard _Tag_ support to direct the `Serializer` how to `Marshal` or `Unmarshal` the data. For example

```go
type MyContext struct {
  Caller string
  TotalTimeout int `pms:"timeout"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb"`
    BatchSize int `pms:"batchsize"`
    DbTimeout int `pms:"timeout"`
    UpdateRevenue bool
    Signer string
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

In combination with [env](https://github.com/codingconcepts/env) this is a great way of centrally adminitrating your configuration but allow override of those using environment variables. For example
```go
type MyContext struct {
  Caller string
  TotalTimeout int `pms:"timeout",env:TOTAL_TIMEOUT"`
  Db struct {
    ConnectString string `pms:"connection, prefix=global/accountingdb", env:DEBUG_DB_CONNECTION`
    BatchSize int `pms:"batchsize"`
    DbTimeout int `pms:"timeout"`
    UpdateRevenue bool
    Signer string
  }
}

var ctx MyContext

s := ssm.NewSsmSerializer("eap", "test-service")
if _, err := s.Unmarshal(&ctx); err != nil  {
  panic()

if err := env.set(&ctx); err != nil  {
  panic()
}
// If we e.g. set the TOTAL_TIMEOUT = 99 in the env for the lambda 
// the ctx.TotalTimeout will be 99 and hence overridden locally
fmt.Printf("got total timeout of %d and connect using %s ...", ctx.TotalTimeout, ctx.Db.ConnectString)
```
