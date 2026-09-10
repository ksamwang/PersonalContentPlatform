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
  metadata: Record<string, unknown>;
  content_hash: string;
  created_at: string;
};
export type Localization = {
  id: string;
  locale: string;
  state: string;
  slug: string;
  translation_status: string;
  current_revision?: Revision;
};
export type ReadinessIssue = { code:string; message:string; severity:"error"|"warning" };
export type Readiness = { ready:boolean; issues:ReadinessIssue[] };
export type Content = {
  id: string;
  type: "article" | "note" | "page";
  default_locale: string;
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
  usage_count: number;
  width?:number;
  height?:number;
  variants?:{recipe:string;width:number;height:number}[];
};
export type AssetUsage={id:string;owner_id:string;owner_title:string;role:string};
export type GeneralSettings = {
  workspace: { name: string; slug: string; default_locale: string; supported_locales: string[]; timezone: string };
  site: { name: string; public_url: string; description: string; about: string; footer: string; rss_enabled: boolean; theme:string; accent_color:string; share_footer:string };
  auth: { password_login_enabled: boolean; passkey_enabled: boolean; session_ttl_hours: number };
  publication: { default_channel: string; auto_publish: boolean };
};
export type StorageSettings = { id?: string; name: string; provider: "filesystem"|"s3"|"r2"|"oss"; endpoint: string; region: string; bucket: string; access_key_mask?: string; secret_key_set?: boolean; base_path: string; active?: boolean };
export type AISettings = { provider: string; base_url: string; api_key_mask?: string; api_key_set?: boolean; model: string; purpose_models: Record<string,string> };
export type EmbeddingSettings={provider:string;base_url:string;api_key_mask?:string;api_key_set?:boolean;model:string;dimensions:number};
export type MediaSettings={purpose:"ocr"|"transcription";provider:string;base_url:string;api_key_mask?:string;api_key_set?:boolean;model:string};
export type SaveAISettings = Pick<AISettings,"provider"|"base_url"|"model"|"purpose_models"> & {api_key?:string};
export type WorkspaceSettings = { general: GeneralSettings; storage: StorageSettings|null; ai: AISettings; embedding:EmbeddingSettings; media:MediaSettings[]; can_edit: boolean };
export type WebhookEndpoint = {id:string;name:string;url:string;secret_set:boolean;enabled:boolean;event_types:string[]};
export type InboxItem = { id:string; kind:"text"|"link"|"image"|"audio"|"file"; raw_text:string; source_url?:string; asset_id?:string; title:string; extracted_text:string; cover_url?:string; processing_state:"idle"|"processing"|"completed"|"failed"; processing_error:string; duplicate_of?:string; state:"pending"|"converted"|"archived"; converted_content_id?:string; created_at:string; updated_at:string };
export type InboxConversion = { inbox_id:string; content_id:string; localization_id:string };
export type CollectionItem = {id:string;object_id:string;title:string;sort_key:number;annotation:string};
export type CollectionSection = {id:string;title:string;sort_key:number;items:CollectionItem[]};
export type Collection = {id:string;workspace_id:string;title:string;slug:string;visibility:"private"|"unlisted"|"public";sections:CollectionSection[]};
export type EntityAlias={id:string;value:string;locale:string};
export type KnowledgeEntity={id:string;type:string;canonical_name:string;description:string;aliases?:EntityAlias[];created_at:string};
export type KnowledgeRelation={id:string;source_id:string;target_id:string;predicate:string;confirmed:boolean;created_at:string};
export type KnowledgeMention={id:string;revision_id:string;entity_id:string;entity_name:string;content_id:string;content_title:string;locale:string;confidence:number;confirmed:boolean;created_at:string};
export type AISuggestion={id:string;target_id:string;target_title:string;kind:string;payload:{text?:string};state:"pending"|"accepted"|"rejected";created_at:string};
export type AIApplyResult={applied:string[];draft_version?:number};
export type AIRun={id:string;purpose:string;provider:string;model:string;status:"running"|"succeeded"|"failed";usage:{input_tokens?:number;output_tokens?:number};error_code:string;started_at:string;completed_at?:string};
export type PublicationRecord={id:string;content_id:string;title:string;locale:string;channel:string;state:string;scheduled_at?:string;published_at?:string;attempts:number;last_error:string;created_at:string};
export type PublicationChannel={channel:string;name:string;status:string;detail:string};
export type ImportReport={format:string;contents:number;localizations:number;assets:number;entities:number;relations:number;conflicts:string[];renamed:string[];applied:boolean};
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

async function requestItems<T>(path: string, init?:RequestInit): Promise<{ items: T[] }> {
  const response = await request<{ items?: T[] | null }>(path,init);
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
  list: (ws: string,filters:Record<string,string>={}) => {
    const params=new URLSearchParams(Object.entries(filters).filter(([,value])=>value));
    return requestItems<Content>(`/v1/workspaces/${ws}/contents${params.size?`?${params}`:""}`);
  },
  content: (ws: string, id: string) =>
    request<Content>(`/v1/workspaces/${ws}/contents/${id}`),
  updateContentProperties: (ws:string,id:string,value:{localization_id:string;slug:string;visibility:string}) =>
    request<void>(`/v1/workspaces/${ws}/contents/${id}`,{method:"PATCH",body:JSON.stringify(value)}),
  search: (ws: string, q: string,locale="") =>
    requestItems<Content>(
      `/v1/workspaces/${ws}/search?q=${encodeURIComponent(q)}${locale?`&locale=${encodeURIComponent(locale)}`:""}`,
    ),
  archiveContent:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/contents/${id}:archive`,{method:"POST"}),
  restoreContent:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/contents/${id}:restore`,{method:"POST"}),
  deleteContent:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/contents/${id}`,{method:"DELETE"}),
  create: (ws: string, input: object) =>
    request<Content>(`/v1/workspaces/${ws}/contents`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
  createLocalization: (ws:string,contentID:string,value:{locale:string;slug:string;source_locale:string}) =>
    request<Content>(`/v1/workspaces/${ws}/contents/${contentID}/localizations`,{method:"POST",body:JSON.stringify(value)}),
  translateLocalization: (ws:string,targetID:string,sourceID:string) =>
    request<Draft>(`/v1/workspaces/${ws}/localizations/${targetID}:translate`,{method:"POST",body:JSON.stringify({source_localization_id:sourceID})}),
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
  readiness: (ws:string,id:string) => request<Readiness>(`/v1/workspaces/${ws}/localizations/${id}/readiness`),
  publish: (ws: string, content_id: string, locale: string) =>
    request<{ publication_id: string }>(`/v1/workspaces/${ws}/publications`, {
      method: "POST",
      body: JSON.stringify({ content_id, locale }),
    }),
  assets: (ws: string) =>
    requestItems<Asset>(`/v1/workspaces/${ws}/assets/`),
  asset:(ws:string,id:string)=>request<Asset>(`/v1/workspaces/${ws}/assets/${id}`),
  assetUsages:(ws:string,id:string)=>requestItems<AssetUsage>(`/v1/workspaces/${ws}/assets/${id}/usages`),
  archiveAsset:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/assets/${id}:archive`,{method:"POST"}),
  replaceAsset:(ws:string,id:string,replacementID:string)=>request<Asset>(`/v1/workspaces/${ws}/assets/${id}:replace`,{method:"POST",body:JSON.stringify({replacement_asset_id:replacementID})}),
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
  saveAI: (ws: string, value: SaveAISettings) => request<AISettings>(`/v1/workspaces/${ws}/settings/ai`, { method: "PUT", body: JSON.stringify(value) }),
  aiModels: (ws:string,value:{provider:string;base_url:string;api_key?:string}) => requestItems<{id:string}>(`/v1/workspaces/${ws}/settings/ai:models`,{method:"POST",body:JSON.stringify(value)}),
  testAI: (ws: string) => request<{ok:boolean}>(`/v1/workspaces/${ws}/settings/ai:test`, { method: "POST" }),
  saveEmbedding:(ws:string,value:EmbeddingSettings&{api_key?:string})=>request<EmbeddingSettings>(`/v1/workspaces/${ws}/settings/embedding`,{method:"PUT",body:JSON.stringify(value)}),
  testEmbedding:(ws:string)=>request<{ok:boolean}>(`/v1/workspaces/${ws}/settings/embedding:test`,{method:"POST"}),
  webhooks: (ws:string)=>requestItems<WebhookEndpoint>(`/v1/workspaces/${ws}/webhooks`),
  createWebhook:(ws:string,value:{name:string;url:string;secret:string;event_types:string[]})=>request<WebhookEndpoint>(`/v1/workspaces/${ws}/webhooks`,{method:"POST",body:JSON.stringify(value)}),
  setWebhookEnabled:(ws:string,id:string,enabled:boolean)=>request<void>(`/v1/workspaces/${ws}/webhooks/${id}`,{method:"PATCH",body:JSON.stringify({enabled})}),
  deleteWebhook:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/webhooks/${id}`,{method:"DELETE"}),
  inbox: (ws:string,state="pending") => requestItems<InboxItem>(`/v1/workspaces/${ws}/inbox/?state=${encodeURIComponent(state)}`),
  saveMedia:(ws:string,purpose:string,value:MediaSettings&{api_key?:string})=>request<MediaSettings>(`/v1/workspaces/${ws}/settings/media/${purpose}`,{method:"PUT",body:JSON.stringify(value)}),
  captureInbox: (ws:string,value:{kind:InboxItem["kind"];raw_text:string;source_url?:string;asset_id?:string}) => request<InboxItem>(`/v1/workspaces/${ws}/inbox/`,{method:"POST",body:JSON.stringify(value)}),
  processInbox:(ws:string,id:string)=>request<InboxItem>(`/v1/workspaces/${ws}/inbox/${id}:process`,{method:"POST"}),
  archiveInbox: (ws:string,id:string) => request<void>(`/v1/workspaces/${ws}/inbox/${id}:archive`,{method:"POST"}),
  convertInbox: (ws:string,id:string,value:{type:Content["type"];locale:string;slug:string;title:string}) => request<InboxConversion>(`/v1/workspaces/${ws}/inbox/${id}:convert`,{method:"POST",body:JSON.stringify(value)}),
  collections: (ws:string) => requestItems<Collection>(`/v1/workspaces/${ws}/collections`),
  collection: (ws:string,id:string) => request<Collection>(`/v1/workspaces/${ws}/collections/${id}`),
  createCollection: (ws:string,value:{title:string;slug:string;visibility:string}) => request<Collection>(`/v1/workspaces/${ws}/collections`,{method:"POST",body:JSON.stringify(value)}),
  updateCollection: (ws:string,id:string,value:{title:string;slug:string;visibility:string}) => request<Collection>(`/v1/workspaces/${ws}/collections/${id}`,{method:"PUT",body:JSON.stringify(value)}),
  addCollectionSection: (ws:string,id:string,title:string) => request<CollectionSection>(`/v1/workspaces/${ws}/collections/${id}/sections`,{method:"POST",body:JSON.stringify({title})}),
  moveCollectionSection: (ws:string,collectionID:string,sectionID:string,direction:-1|1) => request<void>(`/v1/workspaces/${ws}/collections/${collectionID}/sections/${sectionID}:move`,{method:"POST",body:JSON.stringify({direction})}),
  addCollectionItem: (ws:string,sectionID:string,objectID:string) => request<CollectionItem>(`/v1/workspaces/${ws}/collection-sections/${sectionID}/items`,{method:"POST",body:JSON.stringify({object_id:objectID})}),
  moveCollectionItem: (ws:string,itemID:string,direction:-1|1) => request<void>(`/v1/workspaces/${ws}/collection-items/${itemID}:move`,{method:"POST",body:JSON.stringify({direction})}),
  removeCollectionItem: (ws:string,itemID:string) => request<void>(`/v1/workspaces/${ws}/collection-items/${itemID}`,{method:"DELETE"}),
  revisions: (ws:string,localizationID:string) => requestItems<Revision>(`/v1/workspaces/${ws}/localizations/${localizationID}/revisions`),
  revision: (ws:string,localizationID:string,revisionID:string) => request<Revision>(`/v1/workspaces/${ws}/localizations/${localizationID}/revisions/${revisionID}`),
  restoreRevision: (ws:string,localizationID:string,revisionID:string,version:number) => request<Draft>(`/v1/workspaces/${ws}/localizations/${localizationID}/revisions/${revisionID}:restore`,{method:"POST",body:JSON.stringify({version})}),
  preview: (ws:string,localizationID:string) => request<{token:string;expires_at:string}>(`/v1/workspaces/${ws}/localizations/${localizationID}/preview`,{method:"POST"}),
  entities:(ws:string,q="")=>requestItems<KnowledgeEntity>(`/v1/workspaces/${ws}/knowledge/entities?q=${encodeURIComponent(q)}`),
  createEntity:(ws:string,value:{type:string;canonical_name:string;description:string})=>request<KnowledgeEntity>(`/v1/workspaces/${ws}/knowledge/entities`,{method:"POST",body:JSON.stringify(value)}),
  updateEntity:(ws:string,id:string,value:{type:string;canonical_name:string;description:string})=>request<void>(`/v1/workspaces/${ws}/knowledge/entities/${id}`,{method:"PUT",body:JSON.stringify(value)}),
  deleteEntity:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/knowledge/entities/${id}`,{method:"DELETE"}),
  addEntityAlias:(ws:string,id:string,value:{value:string;locale:string})=>request<void>(`/v1/workspaces/${ws}/knowledge/entities/${id}/aliases`,{method:"POST",body:JSON.stringify(value)}),
  deleteEntityAlias:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/knowledge/aliases/${id}`,{method:"DELETE"}),
  knowledgeRelations:(ws:string,id:string)=>requestItems<KnowledgeRelation>(`/v1/workspaces/${ws}/knowledge/objects/${id}/relations`),
  createKnowledgeRelation:(ws:string,value:{source_id:string;target_id:string;predicate:string;confirmed:boolean})=>request<KnowledgeRelation>(`/v1/workspaces/${ws}/knowledge/relations`,{method:"POST",body:JSON.stringify(value)}),
  confirmKnowledgeRelation:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/knowledge/relations/${id}:confirm`,{method:"POST"}),
  deleteKnowledgeRelation:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/knowledge/relations/${id}`,{method:"DELETE"}),
  knowledgeMentions:(ws:string)=>requestItems<KnowledgeMention>(`/v1/workspaces/${ws}/knowledge/mentions`),
  extractKnowledgeMentions:(ws:string)=>requestItems<KnowledgeMention>(`/v1/workspaces/${ws}/knowledge/mentions:extract`,{method:"POST"}),
  confirmKnowledgeMention:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/knowledge/mentions/${id}:confirm`,{method:"POST"}),
  aiSuggest:(ws:string,value:{target_id:string;localization_id:string;purpose:string;input:string})=>request<{suggestion_id:string;state:string}>(`/v1/workspaces/${ws}/ai/suggestions`,{method:"POST",body:JSON.stringify(value)}),
  aiSuggestions:(ws:string,state="")=>requestItems<AISuggestion>(`/v1/workspaces/${ws}/ai/suggestions${state?`?state=${state}`:""}`),
  reviewAISuggestion:(ws:string,id:string,state:"accepted"|"rejected")=>request<AIApplyResult>(`/v1/workspaces/${ws}/ai/suggestions/${id}:review`,{method:"POST",body:JSON.stringify({state})}),
  aiRuns:(ws:string)=>requestItems<AIRun>(`/v1/workspaces/${ws}/ai/runs`),
  rebuildSearchIndex:(ws:string)=>request<{chunks:number}>(`/v1/workspaces/${ws}/search:index`,{method:"POST"}),
  hybridSearch:(ws:string,q:string,locale="")=>requestItems<SearchHit>(`/v1/workspaces/${ws}/hybrid-search?q=${encodeURIComponent(q)}${locale?`&locale=${encodeURIComponent(locale)}`:""}`),
  rag:(ws:string,query:string,locale="")=>request<{answer:string;sources:SearchHit[]}>(`/v1/workspaces/${ws}/rag`,{method:"POST",body:JSON.stringify({query,locale})}),
  publicationRecords:(ws:string)=>requestItems<PublicationRecord>(`/v1/workspaces/${ws}/publication-records`),
  publicationChannels:(ws:string)=>requestItems<PublicationChannel>(`/v1/workspaces/${ws}/publication-channels`),
  schedulePublication:(ws:string,value:{content_id:string;locale:string;scheduled_at:string})=>request<{publication_id:string}>(`/v1/workspaces/${ws}/publication-records:schedule`,{method:"POST",body:JSON.stringify(value)}),
  retryPublication:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/publication-records/${id}:retry`,{method:"POST"}),
  withdrawPublication:(ws:string,id:string)=>request<void>(`/v1/workspaces/${ws}/publication-records/${id}:withdraw`,{method:"POST"}),
  previewImport:async(ws:string,format:string,file:File)=>{const response=await fetch(`/api/v1/workspaces/${ws}/imports/${format}`,{method:"POST",credentials:"include",headers:{"Content-Type":"application/octet-stream"},body:file});if(!response.ok){const problem=await response.json().catch(()=>({}));throw new Error(problem.detail??"导入预检失败")};return response.json() as Promise<ImportReport>},
  applyImport:async(ws:string,format:string,file:File)=>{const response=await fetch(`/api/v1/workspaces/${ws}/imports/${format}?mode=apply`,{method:"POST",credentials:"include",headers:{"Content-Type":"application/octet-stream"},body:file});if(!response.ok){const problem=await response.json().catch(()=>({}));throw new Error(problem.detail??"导入失败")};return response.json() as Promise<ImportReport>},
};
export type SearchHit={content_id:string;title:string;summary:string;locale:string;type:string;slug:string;excerpt:string;score:number};
