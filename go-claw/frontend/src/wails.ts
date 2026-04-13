// This file is the placeholder for Wails JSD bindings.
// Run `wails generate module` or `wails dev` to generate these.

import type { WailsWindow } from './wails.d';

declare const window: Window & WailsWindow;

export interface Tool {
  Name: string;
  Args: Record<string, any>;
}

export function Execute(toolJSON: string): Promise<string> {
  return window.go.main.tools.Execute(toolJSON);
}

export async function ExecuteTool(toolName: string, args: Record<string, any>): Promise<string> {
  const tool: Tool = {
    Name: toolName,
    Args: args
  }
  return await Execute(JSON.stringify(tool))
}

export function GetAppInfo(): Promise<Record<string, string>> {
  return window.go.main.App.GetAppInfo();
}