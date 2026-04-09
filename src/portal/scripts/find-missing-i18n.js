/**
 * Script to find missing i18n translation keys in Harbor portal.
 * Scans HTML and TypeScript files for translation key usage and compares
 * against defined keys in language JSON files.
 * Also verifies all language files have the same keys as en-us.
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
        // Keys with dots (e.g., 'BUTTON.CANCEL' | translate)
        /'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'\s*\|\s*translate/g,
        /"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"\s*\|\s*translate/g,
        // Single word keys (e.g., 'CANCEL' | translate)
        /'([A-Z][A-Z0-9_]{2,})'\s*\|\s*translate/g,
        /"([A-Z][A-Z0-9_]{2,})"\s*\|\s*translate/g,
        // Service calls with dots
        /translate\.(get|instant)\(\s*'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'/g,
        /translate\.(get|instant)\(\s*"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"/g,
        // Service calls single word
        /translate\.(get|instant)\(\s*'([A-Z][A-Z0-9_]{2,})'/g,
        /translate\.(get|instant)\(\s*"([A-Z][A-Z0-9_]{2,})"/g,
        /translateService\.(get|instant)\(\s*'([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)'/gi,
        /translateService\.(get|instant)\(\s*"([A-Z][A-Z0-9_]+\.[A-Z0-9_.]+)"/gi,
        /translateService\.(get|instant)\(\s*'([A-Z][A-Z0-9_]{2,})'/gi,
        /translateService\.(get|instant)\(\s*"([A-Z][A-Z0-9_]{2,})"/gi,
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
 * Get all language files except en-us.
 */
function getOtherLangFiles() {
    const files = fs.readdirSync(LANG_DIR);
    return files.filter(function(f) {
        return f.endsWith('-lang.json') && f !== 'en-us-lang.json';
    });
}

/**
 * Check if all language files have the same keys as en-us.
 */
function checkLangFileSync(enKeys) {
    const langFiles = getOtherLangFiles();
    const issues = {};
    let totalMissing = 0;

    for (let i = 0; i < langFiles.length; i++) {
        const langFile = langFiles[i];
        const langPath = path.join(LANG_DIR, langFile);
        const langData = JSON.parse(fs.readFileSync(langPath, 'utf-8'));
        const langKeys = extractKeys(langData);
        const langKeySet = {};

        for (let j = 0; j < langKeys.length; j++) {
            langKeySet[langKeys[j]] = true;
        }

        const missing = [];
        for (let k = 0; k < enKeys.length; k++) {
            if (!langKeySet[enKeys[k]]) {
                missing.push(enKeys[k]);
            }
        }

        if (missing.length > 0) {
            issues[langFile] = missing;
            totalMissing += missing.length;
        }
    }

    return { issues: issues, totalMissing: totalMissing };
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

    let hasErrors = false;

    // Check for keys used in code but missing from en-us
    console.log('=== MISSING KEYS (used in code but not in en-us) ===\n');
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
        hasErrors = true;
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
        console.log('\nTotal missing from en-us: ' + missingKeys.length + '\n');
    }

    // Check that all language files have the same keys as en-us
    console.log('=== LANGUAGE FILE SYNC CHECK ===\n');
    const syncResult = checkLangFileSync(definedKeys);

    if (syncResult.totalMissing === 0) {
        console.log('All language files are in sync with en-us!\n');
    } else {
        hasErrors = true;
        const langFiles = Object.keys(syncResult.issues);
        langFiles.sort();

        for (let i = 0; i < langFiles.length; i++) {
            const langFile = langFiles[i];
            const missing = syncResult.issues[langFile];
            console.log(langFile + ': ' + missing.length + ' missing keys');
            var maxShow = Math.min(missing.length, 5);
            for (var m = 0; m < maxShow; m++) {
                console.log('    - ' + missing[m]);
            }
            if (missing.length > 5) {
                console.log('    - ... and ' + (missing.length - 5) + ' more');
            }
        }
        console.log('\nTotal keys missing across all language files: ' + syncResult.totalMissing);
    }

    if (hasErrors) {
        process.exitCode = 1;
    }
}

main();
