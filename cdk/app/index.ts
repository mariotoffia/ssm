import path = require('path');
import fs = require('fs');
import { Reporter } from './reporter';
import { Emitter } from './emitter';
import { Template } from './template';

// Handle options
const argv = require('yargs')
  .option('outfile', {
    alias: 'o',
    describe: 'An optional outfile to write the resulting CDK Construct'
  })
  .option('stdout', {
    describe: 'Output the result onto stdout. This may be combined with --outfile'
  })
  .option('infile', {
    alias: 'i',
    describe: 'The ssm report file to read from filesystem instead of default stdin'
  })
  .option('tsconfig', {
    describe: 'Optional tsconfig.json file to use when generating the source code'
  })
  .option('clsname', {
    alias: 'c',
    describe: 'Optional a class name for the generated CDK class'
  })
  .argv

if (argv.help) {
  console.log("usage --stdout --outfile=<filename.ts>");
  process.exit(0);
}

// Reporter
const reporter = new Reporter();

// Get data from file or stdin
if (argv.infile) {
  reporter.file = argv.infile;
} else {
  reporter.data = fs.readFileSync(0, 'utf-8');
}
// Parse the ssm JSON data into a report
reporter.Parse();

// Templates
const pmsTemplate = new Template(path.join(__dirname,"../templates/pms.txt"), true);
const asmTemplate = new Template(path.join(__dirname,"../templates/asm.txt"), true);
const newFileTemplate = new Template(path.join(__dirname,"../templates/newfile.txt"), true);

// Emitter
const emitter = new Emitter(reporter, pmsTemplate, asmTemplate, newFileTemplate);
emitter.outfile = argv.outfile;
emitter.tsconfig = argv.tsconfig;

// Emit the CDK Construct
var result = emitter.Emit();
if (argv.stdout) {
  console.log(result);
}