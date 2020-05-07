import fs = require('fs');

export class Template {
  private readonly _template: string;

  constructor(template: string, isfilepath: boolean) {
    if (isfilepath) {
      this._template = fs.readFileSync(template).toString();
    } else {
      this._template = template;
    }
  }

  public RenderIndexed(obj: any, index: number): string {
    const copy = (JSON.parse(JSON.stringify(obj)));
    if (copy.tags) {
      delete copy.tags;
    }
    const flat = this.Flatten(copy);
    flat["index"] = index;

    if (obj.tags) {
      flat.tags = obj.tags;
    }

    return this.Substitute(this._template, flat, "");
  }

  public RenderIndexedCfnTags(obj: any, index: number): string {
    const copy = (JSON.parse(JSON.stringify(obj)));
    if (copy.tags) {
      delete copy.tags;
    }
    const flat = this.Flatten(copy);
    flat["index"] = index;

    if (obj.tags) {
      var nv = new Array();

      Object.keys(obj.tags).forEach((k) =>  {
        nv.push({key: k, value: obj.tags[k]});
      });
      
      flat.tags = nv;
    }

    return this.Substitute(this._template, flat, "");
  }

  public Render(obj: any): string {
    const flat = this.Flatten(obj)
    const result = this.Substitute(this._template, flat, "")

    return result
  }

  private Flatten<T extends Record<string, any>>(
    object: T,
    path: string | null = null,
    separator = '.'
  ): T {
    return Object.keys(object).reduce((acc: T, key: string): T => {
      const newPath = [path, key].filter(Boolean).join(separator);
      if (typeof object?.[key] === 'object' && object?.[key] == null) {
        return { ...acc, [newPath]: object[key] };
      }

      return typeof object?.[key] === 'object'
        ? {  ...acc, ...this.Flatten(object[key], newPath, separator) }
        : { ...acc, [newPath]: object[key] };
    }, {} as T);
  }

  private Substitute(str: string, data: any, undefval: string) {
    return str.replace(/\{ *([\w|._]+) *\}/g, function (str, key) {
      var value = data[key];
      if (value === undefined) {
        value = undefval;
      } else if (typeof value === 'function') {
        value = value(data);
      } else if (typeof value === 'object') {
        value = JSON.stringify(value);
      }
      return value;
    });
  }
}