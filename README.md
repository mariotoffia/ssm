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

s := NewSsmSerializer("eap", "test-service")
err := s.Unmarshal(&ctx)
if err != nil {
  panic()
}

fmt.Printf("got total timeout of %d and connect using %s ...", ctx.TotalTimeout, ctx.Db.ConnectString)
```

The above example shows how to blend _PMS_ backed data with data set by the service itself to perform the work. Note that the `ConnectString` is a global setting and hence independant on the service it will be retrieved from _/{env}/global/accountingdb/connection_ parameter. In this way it is possible to constrain parameters to a single service, share between services or have notion of global parameters. Environment is *always* present, thus mandatory.

