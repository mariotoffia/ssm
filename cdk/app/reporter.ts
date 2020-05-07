import fs = require('fs');

export interface Parameter {
  type: string;
  name: string;
  keyid: string;
  description: string;
  tags: Map<string, string>;
  value: string
  details: ParameterDetails;
}

export interface ParameterDetails {
}

export interface PmsParameterDetails extends ParameterDetails {
  pattern: string;
  tier: string
}

export interface AsmParameterDetails extends ParameterDetails {
  strkey: string;
}

export interface Report {
  parameters: Parameter[];
}

export class Reporter {
  private _file: string;
  private _data: string;
  private _report: Report;

  get file(): string { return this._file; }
  set file(file: string) { this._file = file }
  get data(): string { return this._data; }
  set data(data:string) { this._data = data;}
  get report(): Report { return this._report }
  
  public Parse() {
    if (this._file) {
      this._data = fs.readFileSync(this._file).toString();
    }

    this._report = <Report>JSON.parse(this._data);
  }
}