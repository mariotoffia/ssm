# Introduction
This package generates CDK classes, using templates, from the go ssm report.

## Reporting
It is possible to generate a report as object and JSON using the golang [ssm report](https://github.com/mariotoffia/ssm). This may be used in a _DevOps_ pipeline to e.g. use CDK to create Parameter Store & Secrets Manager entities to be provisioned using cloud formation. Hence it is possible to have
default values by passing the struct as value instead of `nil` pointer with a specific `interface` type.

For example, the following
```go
type Sample struct {
  ConnectionString string `asm:"connectstring, strkey=password, gurka=biffen, nasse=hunden"`
  Secret string `asm:"mysecret"`
  Parameter string `pms:"parameter, description=A sample value, pattern=.*, my=hobby, by=test"`
}

set := Sample{
  ConnectString: "{\"user\":\"nisse\"}",
  Secret: "{\"private\": \"nobody knows\", \"lockkey\":\"eeej1¤¤&1!\"}",
  Parameter: "a parameter"
}

s := NewSsmSerializer("dev", "test-service")
objs, json, err := s.ReportWithOpts(&set, NoFilter, true)
if err != nil {
  panic(err)
}
```

Renders a _JSON_ report on the following format:
```json
{
  "parameters": [
    {
      "type": "secrets-manager",
      "fqname": "/dev/test-service/connectstring",
      "keyid": "",
      "description": "",
      "tags": {"gurka":"biffen","nasse":"hunden"},
      "details": {
        "strkey": "password"
      },
      "value": "{\"user\": \"nisse\"}"      
    },
    {
      "type": "secrets-manager",
      "fqname": "/dev/test-service/mysecret",
      "keyid": "",
      "description": "",
      "tags": {},
      "details": {
        "strkey": null
      },
      "value": "{\"private\": \"nobody knows\", \"lockkey\":\"eeej1¤¤&1!\"}"      
    },
    {
      "type": "parameter-store",
      "fqname": "/dev/test-service/parameter",
      "keyid": "",
      "description": "A sample value",
      "tags": {"my":"hobby", "by": "test"},
      "details": {
        "pattern": ".*",
        "tier": "Standard"
      },
      "value": "a parameter",
      "valuetype": "String"
    }                
  ]
}
```

However since it returns a `Report` object containing `Parameters` you may do your own _JSON_ or otherwise generation from the `Report` object. It is also possible to use _filter_ to filter out fields in same manner as `Marshal` and `Unmarshal` works.

From this JSON it is possible to transform it to e.g. CDK `@aws-cdk/aws-ssm` object such as
```js
new ssm.StringParameter(stack, 'Parameter', {
  allowedPattern: '.*',
  description: 'The value Foo',
  parameterName: 'FooParameter',
  stringValue: 'Foo',
  tier: ssm.ParameterTier.ADVANCED,
});
```

## CDK Generator
This is an implementation of a CDK generator, that is template driven to fit your CDK generation needs.For example given the report _JSON_ file above and use _ssm-cdk-generator_ will output the following using default templates.

```typescript
import * as cdk from '@aws-cdk/core';
    import * as asm from '@aws-cdk/aws-secretsmanager';
    import * as pms from '@aws-cdk/aws-ssm';

    export class SsmParamsConstruct extends cdk.Construct {
      constructor(scope: cdk.Construct, id: string) {
        super(scope, id);

        // SetupSecrets & SetupParameters must be named exactly as stated below!
        this.SetupSecrets();
        this.SetupParameters();
      }

      private SetupSecrets() {
              new asm.CfnSecret(this, 'Secret0', {
                description: '',
                name: '/dev/test-service/connectstring',
                generateSecretString: {
                  secretStringTemplate: '{"user": "nisse"}',
                  generateStringKey: 'password',
                },
                tags: [{"key":"gurka","value":"biffen"},{"key":"nasse","value":"hunden"}]
              });
              new asm.CfnSecret(this, 'Secret1', {
                description: '',
                name: '/dev/test-service/mysecret',
                secretString: '{"private": "nobody knows", "lockkey":"eeej1¤¤&1!"}',
                tags: []
              });

      }

      private SetupParameters() {
          new pms.CfnParameter(this, 'Parameter0', {
                name: '/dev/test-service/parameter',
                type: 'String',
                value: 'a parameter',
                allowedPattern: '.*',
                description: 'A sample value',
                policies: ''
                tags: {"my":"hobby","by":"test"},
                tier: 'Standard'
              });
      }
    }
```

Use the _node app/index.js --help_ to get help on which parameters you may use.

```bash
Options:
  --help           Show help                                           [boolean]
  --version        Show version number                                 [boolean]
  --outfile, -o    An optional outfile to write the resulting CDK Construct
  --stdout         Output the result onto stdout. This may be combined with
                   --outfile
  --infile, -i     The ssm report file to read from filesystem instead of
                   default stdin
  --tsconfig       Optional tsconfig.json file to use when generating the source
                   code
  --clsname, -c    Optional a class name for the generated CDK class
  --tmplasm, --ta  Optional a template fqfilepath that shall be used for asm
                   parameter
  --tmplpms, --tp  Optional a template fq filepath that shall be used for
                   generating pms parameter
  --tmplclz, --tc  Optional a template fq filepath that shall be used for
                   generating a new file / class
```

Templates are very simple and the default resides under the _ssm-cdk-generator/templates for reviewing. For example a template to generate a single Parameter Store Parameter could look like this.

```typescript
new pms.CfnParameter(this, 'Parameter{index}', {
      name: '{fqname}',
      type: '{valuetype}',
      value: '{value}',
      allowedPattern: '{details.pattern}',
      description: '{description}',
      policies: '{policies}'
      tags: {tags},
      tier: '{details.tier}'
    });
```