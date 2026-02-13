const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

const projectRoot = path.resolve(__dirname, '..');
const srcDir = path.join(projectRoot, 'src');
const outDir = path.join(projectRoot, 'dist');

function ensureDir(dirPath) {
  fs.mkdirSync(dirPath, { recursive: true });
}

function walk(dirPath, fileList = []) {
  const entries = fs.readdirSync(dirPath, { withFileTypes: true });

  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      walk(fullPath, fileList);
      continue;
    }

    fileList.push(fullPath);
  }

  return fileList;
}

function writeFileWithMap(targetJsFile, outputText, sourceMapText) {
  ensureDir(path.dirname(targetJsFile));
  const mapFileName = `${path.basename(targetJsFile)}.map`;
  const outputWithMap = `${outputText}\n//# sourceMappingURL=${mapFileName}`;
  fs.writeFileSync(targetJsFile, outputWithMap, 'utf8');
  fs.writeFileSync(`${targetJsFile}.map`, sourceMapText, 'utf8');
}

function transpileTypeScriptFile(sourceFilePath) {
  const sourceText = fs.readFileSync(sourceFilePath, 'utf8');
  const relativePath = path.relative(srcDir, sourceFilePath);
  const outputFilePath = path.join(outDir, relativePath.replace(/\.ts$/i, '.js'));

  const result = ts.transpileModule(sourceText, {
    fileName: sourceFilePath,
    reportDiagnostics: true,
    compilerOptions: {
      target: ts.ScriptTarget.ES2022,
      module: ts.ModuleKind.CommonJS,
      moduleResolution: ts.ModuleResolutionKind.NodeJs,
      sourceMap: true,
      esModuleInterop: true,
      allowSyntheticDefaultImports: true,
      resolveJsonModule: true,
      experimentalDecorators: true,
      emitDecoratorMetadata: true,
      noEmitHelpers: false,
      importHelpers: false,
    },
  });

  const diagnostics = result.diagnostics || [];
  if (diagnostics.length > 0) {
    const formatted = ts.formatDiagnosticsWithColorAndContext(diagnostics, {
      getCurrentDirectory: () => projectRoot,
      getCanonicalFileName: (fileName) => fileName,
      getNewLine: () => '\n',
    });

    process.stderr.write(formatted);
    process.exitCode = 1;
    return;
  }

  writeFileWithMap(outputFilePath, result.outputText, result.sourceMapText || '{}');
}

function copyJsonFile(sourceFilePath) {
  const relativePath = path.relative(srcDir, sourceFilePath);
  const targetPath = path.join(outDir, relativePath);
  ensureDir(path.dirname(targetPath));
  fs.copyFileSync(sourceFilePath, targetPath);
}

function main() {
  fs.rmSync(outDir, { recursive: true, force: true });
  ensureDir(outDir);

  const allFiles = walk(srcDir);

  for (const filePath of allFiles) {
    if (filePath.endsWith('.test.ts')) {
      continue;
    }

    if (filePath.endsWith('.ts')) {
      transpileTypeScriptFile(filePath);
      continue;
    }

    if (filePath.endsWith('.json')) {
      copyJsonFile(filePath);
    }
  }

  if (process.exitCode && process.exitCode !== 0) {
    process.exit(process.exitCode);
  }

  process.stdout.write('Build completed: transpiled TypeScript to dist\n');
}

main();
