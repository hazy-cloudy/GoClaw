export interface WailsWindow {
  go: {
    main: {
      tools: {
        Execute(toolJSON: string): Promise<string>;
      };
      App: {
        GetAppInfo(): Promise<Record<string, string>>;
        GetEmotion(): Promise<string>;
        LoadConfig(): Promise<string>;
        SaveConfig(jsonStr: string): Promise<boolean>;
        LoadInitSettings(): Promise<string>;
        SaveInitSettings(jsonStr: string): Promise<boolean>;
        ResetInitSettings(): Promise<boolean>;
        SetClickThrough(enabled: boolean): Promise<void>;
      };
    };
  };
  runtime: {
    EventsOn(eventName: string, callback: (data: any) => void): void;
    EventsOff(eventName: string): void;
    EventsEmit(eventName: string, data?: any): void;
    WindowSetSize(width: number, height: number): void;
    WindowSetPosition(x: number, y: number): void;
    WindowGetPosition(): Promise<{ x: number; y: number }>;
    WindowShow(): void;
    WindowHide(): void;
  };
}

declare global {
  interface Window extends WailsWindow {}
}

export {};