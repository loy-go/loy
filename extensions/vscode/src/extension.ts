import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as path from 'path';

interface LoyDiagnostic {
  severity: string;
  code: string;
  message: string;
  file?: string;
  line?: number;
  column?: number;
  hint?: string;
}

export function activate(context: vscode.ExtensionContext) {
  const diagnosticCollection = vscode.languages.createDiagnosticCollection('loy');
  context.subscriptions.push(diagnosticCollection);

  const outputChannel = vscode.window.createOutputChannel('Loy');
  context.subscriptions.push(outputChannel);

  // 1. Run Check Command
  const runCheck = async () => {
    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (!workspaceFolders || workspaceFolders.length === 0) {
      return;
    }

    if (!vscode.workspace.isTrusted) {
      outputChannel.appendLine('[Security] Skipping loy check in untrusted workspace.');
      return;
    }

    const rootDir = workspaceFolders[0].uri.fsPath;
    const config = vscode.workspace.getConfiguration('loy');
    const loyPath = config.get<string>('path') || 'loy';
    const deep = config.get<boolean>('deepAnalysis') || false;

    const args = ['check', '--format', 'json'];
    if (deep) {
      args.push('--deep');
    }

    cp.execFile(loyPath, args, { cwd: rootDir }, (error, stdout, stderr) => {
      diagnosticCollection.clear();
      const fileDiagnosticsMap = new Map<string, vscode.Diagnostic[]>();

      if (!stdout && stderr) {
        outputChannel.appendLine(`[Error] ${stderr}`);
        return;
      }

      try {
        const diagnostics: LoyDiagnostic[] = JSON.parse(stdout || '[]');
        for (const diag of diagnostics) {
          if (!diag.file) {
            continue;
          }

          const filePath = path.isAbsolute(diag.file) ? diag.file : path.join(rootDir, diag.file);
          const line = Math.max(0, (diag.line || 1) - 1);
          const range = new vscode.Range(line, 0, line, 200);

          let severity = vscode.DiagnosticSeverity.Error;
          if (diag.severity === 'warning') {
            severity = vscode.DiagnosticSeverity.Warning;
          } else if (diag.severity === 'info') {
            severity = vscode.DiagnosticSeverity.Information;
          }

          let msg = diag.message;
          if (diag.hint) {
            msg += `\nHint: ${diag.hint}`;
          }

          const vsDiag = new vscode.Diagnostic(range, msg, severity);
          vsDiag.code = diag.code;
          vsDiag.source = 'Loy Architecture';

          const existing = fileDiagnosticsMap.get(filePath) || [];
          existing.push(vsDiag);
          fileDiagnosticsMap.set(filePath, existing);
        }

        for (const [filePath, diags] of fileDiagnosticsMap.entries()) {
          diagnosticCollection.set(vscode.Uri.file(filePath), diags);
        }

        if (diagnostics.length === 0) {
          outputChannel.appendLine('✓ All architecture rules passed.');
        } else {
          outputChannel.appendLine(`⚠️ Found ${diagnostics.length} architectural violation(s).`);
        }
      } catch (err) {
        outputChannel.appendLine(`Failed to parse loy check output: ${err}`);
      }
    });
  };

  // 2. Register Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('loy.check', runCheck)
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('loy.routes', () => {
      const term = vscode.window.createTerminal('Loy Routes');
      term.show();
      term.sendText('loy routes');
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('loy.graph', () => {
      const term = vscode.window.createTerminal('Loy Architecture Graph');
      term.show();
      term.sendText('loy graph');
    })
  );

  context.subscriptions.push(
    vscode.commands.registerCommand('loy.dev', () => {
      const term = vscode.window.createTerminal('Loy Dev');
      term.show();
      term.sendText('loy dev');
    })
  );

  // 3. Document Save Hook
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      if (doc.languageId === 'go') {
        const config = vscode.workspace.getConfiguration('loy');
        if (config.get<boolean>('checkOnSave')) {
          runCheck();
        }
      }
    })
  );

  // 4. Quick Fix Code Action Provider
  context.subscriptions.push(
    vscode.languages.registerCodeActionsProvider(
      'go',
      {
        provideCodeActions(document, range, context) {
          const actions: vscode.CodeAction[] = [];
          for (const diag of context.diagnostics) {
            if (diag.code === 'LOY-ARCH-002' || diag.code === 'ARCH-002') {
              const fix = new vscode.CodeAction(
                'Invert dependency via Domain Repository Interface (Loy Clean Architecture)',
                vscode.CodeActionKind.QuickFix
              );
              fix.diagnostics = [diag];
              actions.push(fix);
            } else if (diag.code === 'LOY-ARCH-009' || diag.code === 'ARCH-009') {
              const fix = new vscode.CodeAction(
                'Remove web framework import from Application Service',
                vscode.CodeActionKind.QuickFix
              );
              fix.diagnostics = [diag];
              actions.push(fix);
            }
          }
          return actions;
        }
      },
      {
        providedCodeActionKinds: [vscode.CodeActionKind.QuickFix]
      }
    )
  );
}

export function deactivate() {}
