// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

export interface FileFilterValue {
    include?: string[];
    exclude?: string[];
}

export function splitPatterns(input: string): string[] {
    return (input || '')
        .split(/[\s,]+/)
        .map(p => p.trim())
        .filter(p => p.length > 0);
}

// parseFileFilter accepts both value forms of the "file" replication filter:
// the object {include, exclude} and the string "*.json, !onnx/**".
export function parseFileFilter(value: any): FileFilterValue {
    if (!value) {
        return { include: [], exclude: [] };
    }
    if (typeof value === 'string') {
        const include: string[] = [];
        const exclude: string[] = [];
        splitPatterns(value).forEach(p =>
            p.startsWith('!') ? exclude.push(p.slice(1)) : include.push(p)
        );
        return { include, exclude };
    }
    return { include: value.include || [], exclude: value.exclude || [] };
}

export function isEmptyFileFilter(f: FileFilterValue): boolean {
    return !f || (!f.include?.length && !f.exclude?.length);
}

// Mirrors pkg/reg/model.FileFilter.Match: a pattern without "/" matches the
// base name at any depth, other patterns the full path (doublestar).
export function matchFileFilter(f: FileFilterValue, path: string): boolean {
    if (isEmptyFileFilter(f)) {
        return true;
    }
    const included = !f.include?.length || matchAny(f.include, path);
    return included && !matchAny(f.exclude || [], path);
}

function matchAny(patterns: string[], path: string): boolean {
    const base = path.split('/').pop();
    return patterns.some(p =>
        globToRegExp(p).test(p.includes('/') ? path : base)
    );
}

export function globToRegExp(glob: string): RegExp {
    let re = '';
    for (let i = 0; i < glob.length; i++) {
        const c = glob[i];
        if (c === '*') {
            if (glob[i + 1] === '*') {
                i++;
                if (glob[i + 1] === '/') {
                    i++;
                    re += '(?:.*/)?';
                } else {
                    re += '.*';
                }
            } else {
                re += '[^/]*';
            }
        } else if (c === '?') {
            re += '[^/]';
        } else if (c === '{') {
            const end = glob.indexOf('}', i);
            if (end < 0) {
                re += '\\{';
                continue;
            }
            re +=
                '(?:' +
                glob
                    .slice(i + 1, end)
                    .split(',')
                    .map(alt => globToRegExp(alt).source.slice(1, -1))
                    .join('|') +
                ')';
            i = end;
        } else if (c === '[') {
            const end = glob.indexOf(']', i);
            if (end < 0) {
                re += '\\[';
                continue;
            }
            let cls = glob.slice(i + 1, end);
            if (cls.startsWith('!') || cls.startsWith('^')) {
                cls = '^' + cls.slice(1);
            }
            re += '[' + cls.replace(/\\/g, '\\\\') + ']';
            i = end;
        } else if ('.+^$()|\\/'.includes(c)) {
            re += '\\' + c;
        } else {
            re += c;
        }
    }
    return new RegExp(`^${re}$`);
}

// The replication name and tag filters treat these characters as pattern
// syntax; a value without them names exactly one repository or tag.
export function isExactPattern(value: string): boolean {
    return !!value && !/[*?[\]{},]/.test(value);
}

const CLASS_ICONS: { [cls: string]: string } = {
    weight: 'layers',
    config: 'cog',
    code: 'code',
    doc: 'note',
};

const NAME_RULES: { icon: string; test: RegExp }[] = [
    {
        icon: 'certificate',
        test: /^(license|licence|notice|copying)(\.[a-z0-9]+)?$/i,
    },
    { icon: 'note', test: /^(readme.*|.*\.(md|markdown|rst|txt|pdf|adoc))$/i },
    {
        icon: 'layers',
        test: /\.(safetensors|bin|gguf|ggml|pt|pth|ckpt|onnx|h5|hdf5|pb|tflite|msgpack|npy|npz|pkl|mlmodel|engine)(\.[0-9]+)?$/i,
    },
    {
        icon: 'cog',
        test: /^((config|generation_config|tokenizer|tokenizer_config|special_tokens_map|vocab|merges|params|chat_template).*|.*\.(json|jsonl|yaml|yml|toml|ini|cfg|model|tiktoken|jinja))$/i,
    },
    {
        icon: 'code',
        test: /\.(py|ipynb|sh|js|ts|go|rs|c|cpp|h|java|cu|lua|rb)$/i,
    },
    { icon: 'table', test: /\.(csv|tsv|parquet|arrow|feather)$/i },
    { icon: 'file-zip', test: /\.(zip|tar|gz|tgz|bz2|xz|7z|zst)$/i },
    { icon: 'image', test: /\.(png|jpe?g|gif|svg|webp|bmp)$/i },
];

// fileIcon returns the Clarity icon of a model file, from the class the
// source reported when there is one, otherwise from the file name.
export function fileIcon(path: string, cls?: string): string {
    if (cls && CLASS_ICONS[cls]) {
        return CLASS_ICONS[cls];
    }
    const base = (path || '').split('/').pop() || '';
    for (const rule of NAME_RULES) {
        if (rule.test.test(base)) {
            return rule.icon;
        }
    }
    return 'file';
}
