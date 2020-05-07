import { Reporter, Parameter, AsmParameterDetails } from './reporter';
import { Project, StructureKind, SourceFile, ParameterDeclaration } from "ts-morph";
import fs = require('fs');
import { Template } from './template';


// https://github.com/dsherret/ts-morph/tree/latest/packages/ts-morph

export class Emitter {
  private _project: Project;
  private _classname: string = "SsmParamsConstruct";
  private _asmcount: number = 0;
  private _pmscount: number = 0;
  private _sourcefile: SourceFile;
  private _tsconfig: string;
  private _outfile: string;

  constructor(private _reporter: Reporter,
    private readonly _pmsTemplate: Template,
    private readonly _asmTemplate: Template,
    private readonly _asmgkTemplate: Template,
    private readonly _newFileTemplate: Template, ) {
  }

  get reporter(): Reporter { return this._reporter; }
  get classname(): string { return this._classname; }
  set classname(clz: string) { this._classname = clz; }
  // If you initialize with a tsconfig.json, then it will 
  // automatically populate the project with the associated source files.
  // Read more: https://ts-morph.com/setup/
  set tsconfig(fqpath: string) { this._tsconfig = fqpath; }
  get tsconfig(): string { return this._tsconfig; }
  set outfile(fqpath: string) { this._outfile = fqpath; }
  get outfile(): string { return this._outfile; }

  public Emit(): string {
    this.Initialize();

    const str = JSON.stringify(this._reporter.report, null, 2);

    this._reporter.report.parameters.forEach((prm) => {
      switch (prm.type) {
        case "parameter-store":
          this.GeneratePmsParameter(prm);
          break;
        case "secrets-manager":
          this.GenerateAsmParameter(prm);
          break;
        default:
          console.log("WARN: unknown parameter type ", prm);
          process.exit(-1);
      }
    });

    this._project.saveSync();
    var result = this._sourcefile.getFullText();

    if (this.outfile) {
      fs.writeFileSync(this.outfile, result);
    }

    return result;
  }

  private Initialize() {
    this._project = new Project({
      useInMemoryFileSystem: true,
      tsConfigFilePath: this._tsconfig
    });

    this._sourcefile = this._project.createSourceFile(
      `src/${this._classname}.ts`, this._newFileTemplate.Render({ "classname": this._classname }));
  }

  private GeneratePmsParameter(param: Parameter) {
    const cls = this._sourcefile.getClasses()[0];
    const func = cls.getMethodOrThrow("SetupParameters")
    func.addStatements(this._pmsTemplate.RenderIndexed(param, this._pmscount++));
  }

  private GenerateAsmParameter(param: Parameter) {
    const cls = this._sourcefile.getClasses()[0];
    const func = cls.getMethodOrThrow("SetupSecrets")

    var details = <AsmParameterDetails>param.details;
    if (details.strkey) {
      func.addStatements(this._asmgkTemplate.RenderIndexedCfnTags(param, this._asmcount++));  
    } else {
      func.addStatements(this._asmTemplate.RenderIndexedCfnTags(param, this._asmcount++));
    }
  }
}