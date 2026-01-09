/**
 * Script to find missing i18n translation keys in Harbor portal.
 * Scans HTML and TypeScript files for translation key usage and compares
 * against defined keys in language JSON files.
 *
 * Usage: node scripts/find-missing-i18n.js
 */

const fs = require('fs');
const path = require('path');

const PORTAL_DIR = path.join(__dirname, '..');
const LANG_DIR = path.join(PORTAL_DIR, 'src/i18n/lang');
const APP_DIR = path.join(PORTAL_DIR, 'src/app');
const EN_FILE = path.join(LANG_DIR, 'en-us-lang.json');

/**
 * Recursively extract all keys from nested JSON object.
 */
function extractKeys(obj, prefix) {
    prefix = prefix || '';
    const keys = [];
    for (const key of Object.keys(obj)) {
        const value = obj[key];
        const fullKey = prefix ? prefix + '.' + key : key;
        if (typeof value === 'object' && value !== null) {
            keys.push.apply(keys, extractKeys(value, fullKey));
        } else {
            keys.push(fullKey);
        }
    }
    return keys;
}

/**
 * Recursively get all files matching extensions in a directory.
 */
function getFiles(dir, extensions) {
    const files = [];
    const items = fs.readdirSync(dir, { withFileTypes: true });

    for (let i = 0; i < items.length; i++) {
        const item = items[i];
        const fullPath = path.join(dir, item.name);
        if (item.isDirectory()) {
            files.push.apply(files, getFiles(fullPath, extensions));
        } else {
            for (let j = 0; j < extensions.length; j++) {
                if (item.name.endsWith(extensions[j])) {
                    files.push(fullPath);
                    break;
                }
            }
        }
    }
    return files;
}

/**
 * Find all translation keys used in source files.
 */
function findUsedKeys() {
    const usedKeys = {};
    const patterns = [
        /'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'\s*\|\s*translate/g,
        /"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"\s*\|\s*translate/g,
        /translate\.(get|instant)\(\s*'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'/g,
        /translate\.(get|instant)\(\s*"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"/g,
        /translateService\.(get|instant)\(\s*'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'/gi,
        /translateService\.(get|instant)\(\s*"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"/gi,
    ];

    const files = getFiles(APP_DIR, ['.html', '.ts']);

    for (let f = 0; f < files.length; f++) {
        const file = files[f];
        const content = fs.readFileSync(file, 'utf-8');
        const relativePath = path.relative(PORTAL_DIR, file);

        for (let p = 0; p < patterns.length; p++) {
            const pattern = patterns[p];
            let match;
            const regex = new RegExp(pattern.source, pattern.flags);
            while ((match = regex.exec(content)) !== null) {
                const key = match[2] || match[1];
                if (key) {
                    if (!usedKeys[key]) {
                        usedKeys[key] = [];
                    }
                    if (usedKeys[key].indexOf(relativePath) === -1) {
                        usedKeys[key].push(relativePath);
                    }
                }
            }
        }
    }
    return usedKeys;
}

/**
 * Main function
 */
function main() {
    console.log('=== Harbor i18n Missing Keys Finder ===\n');

    if (!fs.existsSync(EN_FILE)) {
        console.error('Error: ' + EN_FILE + ' not found');
        process.exitCode = 1;
        return;
    }

    const enLang = JSON.parse(fs.readFileSync(EN_FILE, 'utf-8'));
    const definedKeys = extractKeys(enLang);
    const definedSet = {};
    for (let i = 0; i < definedKeys.length; i++) {
        definedSet[definedKeys[i]] = true;
    }

    console.log('Keys defined in en-us-lang.json: ' + definedKeys.length);

    const usedKeys = findUsedKeys();
    const usedKeysList = Object.keys(usedKeys);
    console.log('Unique keys used in source code: ' + usedKeysList.length + '\n');

    console.log('=== MISSING KEYS (used in code but not in translation files) ===\n');
    const missingKeys = [];

    for (let i = 0; i < usedKeysList.length; i++) {
        const key = usedKeysList[i];
        if (!definedSet[key]) {
            missingKeys.push({ key: key, files: usedKeys[key] });
        }
    }

    if (missingKeys.length === 0) {
        console.log('No missing keys found!\n');
    } else {
        missingKeys.sort(function(a, b) { return a.key.localeCompare(b.key); });
        for (let i = 0; i < missingKeys.length; i++) {
            var item = missingKeys[i];
            console.log('MISSING: ' + item.key);
            var maxFiles = Math.min(item.files.length, 3);
            for (var j = 0; j < maxFiles; j++) {
                console.log('    -> ' + item.files[j]);
            }
            if (item.files.length > 3) {
                console.log('    -> ... and ' + (item.files.length - 3) + ' more files');
            }
        }
        console.log('\nTotal missing keys: ' + missingKeys.length);
    }

    if (missingKeys.length > 0) {
        process.exitCode = 1;
    }
}

main();
