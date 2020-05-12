package support

// ASMEngine specifies what engine type the JSON
// structure in the secret string is refering to
type ASMEngine string

const (
	// PostgresDBEngine specifies engine type of Postgres SQL
	// of which is by far superior other RDMBSes out there! :)
	PostgresDBEngine ASMEngine = "postgres"
	// SQLServerEngine specifies engine type: SQLServer
	SQLServerEngine ASMEngine = "sqlserver"
	// MariaDBEngine specifies engine type: MariaDB
	MariaDBEngine ASMEngine = "mariadb"
	// MySQLEngine specifies engine type: MySQL
	MySQLEngine ASMEngine = "mysql"
	// OracleEngine specifies engine type: Oracle
	OracleEngine ASMEngine = "oracle"
	// MongoDBEngine specifies engine type: MongoDB
	MongoDBEngine ASMEngine = "mongo"
	// RedShiftEngine specifies engine type: AWS Redshift
	RedShiftEngine ASMEngine = "redshift"
)

// SecretsManagerBaseTemplate is the basis for all managed
// template that may be rotated
type SecretsManagerBaseTemplate struct {
	// Engine is required
	Engine ASMEngine `json:"engine"`
	// Host is required: instance host name/resolvable DNS name
	Host string `json:"host"`
	// Username is required: username
	Username string `json:"username"`
	// Password is required: password. If you provision the
	// secret through cloudformation template, this property
	// must be omitted but set the strkey=password in order
	// for cloud formation to auto generate a password upon
	// provisioning.
	Password string `json:"password,omitempty"`
	// DbName is optional, will default to None if missing
	DbName string `json:"dbname,omitempty"`
	// Port is optional, will default to 3306, 1521, 5432, 1433
	// depending on which database used.
	Port string `json:"port,omitempty"`
}

// SecretsManagerRDSMariaDBRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRDSMariaDBRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRDSMariaDBRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRDSMariaDBRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerRDSMySQLRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRDSMySQLRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRDSMySQLRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRDSMySQLRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerRDSOracleRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRDSOracleRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRDSOracleRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRDSOracleRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerRDSPostgreSQLRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRDSPostgreSQLRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRDSPostgreSQLRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRDSPostgreSQLRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerRDSSQLServerRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRDSSQLServerRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRDSSQLServerRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRDSSQLServerRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerMongoDBRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerMongoDBRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerMongoDBRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerMongoDBRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}

// SecretsManagerRedShiftRotationSingleUser is a template
// to handle single user with rotation in secrets manager
type SecretsManagerRedShiftRotationSingleUser struct {
	SecretsManagerBaseTemplate
}

// SecretsManagerRedShiftRotationMultiUser is a template
// to handle multi user with rotation in secrets manager
type SecretsManagerRedShiftRotationMultiUser struct {
	SecretsManagerBaseTemplate
	// MasterArn required: the ARN of the master secret used
	// to create 2nd user and change passwords
	MasterArn string `json:"masterarn"`
}
