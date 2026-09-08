export type Principal = {
  UserID: string;
  WorkspaceID: string;
  Email: string;
  DisplayName: string;
  Role: string;
};
export type Revision = {
  id: string;
  seq: number;
  title: string;
  summary: string;
  body: Record<string, unknown>;
};
export type Localization = {
  id: string;
  locale: string;
  state: string;
  slug: string;
  current_revision?: Revision;
};
export type Content = {
  id: string;
  type: "article" | "note" | "page";
  visibility: string;
  updated_at: string;
  localizations: Localization[];
};
export type Draft = {
  localization_id: string;
  version: number;
  title: string;
  summary: string;
  body: Record<string, unknown>;
  metadata: Record<string, unknown>;
};
export type SaveDraftInput = Pick<
  Draft,
  "version" | "title" | "summary" | "body" | "metadata"
>;
export type Asset = {
  id: string;
  filename: string;
  media_type: string;
  state: string;
  mime: string;
  size: number;
  sha256: string;
};
export type GeneralSettings = {
  workspace: { name: string; slug: string; default_locale: string; supported_locales: string[]; timezone: string };
  site: { name: string; public_url: string; description: string; about: string; footer: string; rss_enabled: boolean };
  auth: { password_login_enabled: boolean; passkey_enabled: boolean; session_ttl_hours: number };
  publication: { default_channel: string; auto_publish: boolean };
};
export type StorageSettings = { id?: string; name: string; provider: "filesystem"|"s3"|"r2"|"oss"; endpoint: string; region: string; bucket: string; access_key_mask?: string; secret_key_set?: boolean; base_path: string; active?: boolean };
export type AISettings = { provider: string; base_url: string; api_key_mask?: string; api_key_set?: boolean; model: string; purpose_models: Record<string,string> };
export type WorkspaceSettings = { general: GeneralSettings; storage: StorageSettings|null; ai: AISettings; can_edit: boolean };
export class APIError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!response.ok) {
    const problem = await response.json().catch(() => ({}));
    throw new APIError(
      response.status,
      problem.code ?? "request_failed",
      problem.detail ?? "请求失败",
    );
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}

async function requestItems<T>(path: string): Promise<{ items: T[] }> {
  const response = await request<{ items?: T[] | null }>(path);
  return { items: Array.isArray(response.items) ? response.items : [] };
}

export const api = {
  me: () => request<Principal>("/v1/auth/me"),
  setup: (input: object) =>
    request<Principal>("/v1/auth/setup", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  login: (input: object) =>
    request<Principal>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  logout: () => request<void>("/v1/auth/logout", { method: "POST" }),
  list: (ws: string) =>
    requestItems<Content>(`/v1/workspaces/${ws}/contents`),
  content: (ws: string, id: string) =>
    request<Content>(`/v1/workspaces/${ws}/contents/${id}`),
  search: (ws: string, q: string) =>
    requestItems<Content>(
      `/v1/workspaces/${ws}/search?q=${encodeURIComponent(q)}`,
    ),
  create: (ws: string, input: object) =>
    request<Content>(`/v1/workspaces/${ws}/contents`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
  draft: (ws: string, id: string) =>
    request<Draft>(`/v1/workspaces/${ws}/localizations/${id}/draft`),
  saveDraft: (ws: string, id: string, input: SaveDraftInput) =>
    request<Draft>(`/v1/workspaces/${ws}/localizations/${id}/draft`, {
      method: "PUT",
      body: JSON.stringify({
        version: input.version,
        title: input.title,
        summary: input.summary,
        body: input.body,
        metadata: input.metadata,
      }),
    }),
  seal: (ws: string, id: string, version: number) =>
    request<Revision>(`/v1/workspaces/${ws}/localizations/${id}/revisions`, {
      method: "POST",
      body: JSON.stringify({ version }),
    }),
  ready: (ws: string, id: string) =>
    request<void>(`/v1/workspaces/${ws}/localizations/${id}:mark-ready`, {
      method: "POST",
    }),
  publish: (ws: string, content_id: string, locale: string) =>
    request<{ publication_id: string }>(`/v1/workspaces/${ws}/publications`, {
      method: "POST",
      body: JSON.stringify({ content_id, locale }),
    }),
  assets: (ws: string) =>
    requestItems<Asset>(`/v1/workspaces/${ws}/assets/`),
  uploadAsset: async (ws: string, file: File) => {
    const plan = await request<{
      upload_id: string;
      url: string;
      headers: Record<string, string>;
    }>(`/v1/workspaces/${ws}/assets:prepare-upload`, {
      method: "POST",
      body: JSON.stringify({
        filename: file.name,
        mime: file.type,
        size: file.size,
      }),
    });
    const uploadURL = plan.url.startsWith("http")
      ? plan.url
      : `/api${plan.url}`;
    const uploaded = await fetch(uploadURL, {
      method: "PUT",
      headers: plan.headers,
      body: file,
      credentials: "include",
    });
    if (!uploaded.ok) throw new Error("文件上传失败");
    return request<Asset>(`/v1/workspaces/${ws}/assets:finalize-upload`, {
      method: "POST",
      body: JSON.stringify({ upload_id: plan.upload_id }),
    });
  },
  settings: (ws: string) => request<WorkspaceSettings>(`/v1/workspaces/${ws}/settings`),
  saveSettings: <K extends keyof GeneralSettings>(ws: string, section: K, value: GeneralSettings[K]) => request<GeneralSettings>(`/v1/workspaces/${ws}/settings/${section}`, { method: "PUT", body: JSON.stringify(value) }),
  saveStorage: (ws: string, value: StorageSettings & { access_key?: string; secret_key?: string }) => request<StorageSettings>(`/v1/workspaces/${ws}/settings/storage`, { method: "PUT", body: JSON.stringify(value) }),
  testStorage: (ws: string) => request<{ok:boolean}>(`/v1/workspaces/${ws}/settings/storage:test`, { method: "POST" }),
  saveAI: (ws: string, value: AISettings & { api_key?: string }) => request<AISettings>(`/v1/workspaces/${ws}/settings/ai`, { method: "PUT", body: JSON.stringify(value) }),
  testAI: (ws: string) => request<{ok:boolean}>(`/v1/workspaces/${ws}/settings/ai:test`, { method: "POST" }),
};
