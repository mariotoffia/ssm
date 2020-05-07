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
  .option('tmplasm', {
    alias: 'ta',
    describe: 'Optional a template fqfilepath that shall be used for asm parameter'
  })
  .option('tmplpms', {
    alias: 'tp',
    describe: 'Optional a template fq filepath that shall be used for generating pms parameter'
  })
  .option('tmplclz', {
    alias: 'tc',
    describe: 'Optional a template fq filepath that shall be used for generating a new file / class'
  })
  .argv

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
const tmplpms = argv.tmplpms ? argv.tmplpms : path.join(__dirname,"../templates/pms.txt");
const tmplasm = argv.tmplasm ? argv.tmplasm : path.join(__dirname,"../templates/asm.txt");
const tmplclz = argv.tmplclz ? argv.tmplclz : path.join(__dirname,"../templates/newfile.txt");

const pmsTemplate = new Template(tmplpms, true);
const asmTemplate = new Template(tmplasm, true);
const newFileTemplate = new Template(tmplclz, true);

// Emitter
const emitter = new Emitter(reporter, pmsTemplate, asmTemplate, newFileTemplate);
emitter.outfile = argv.outfile;
emitter.tsconfig = argv.tsconfig;

// Emit the CDK Construct
var result = emitter.Emit();
if (argv.stdout) {
  console.log(result);
}