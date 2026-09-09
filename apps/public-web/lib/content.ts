export type PublishedPage = {
  content_id:string;
  locale: string;
  type: string;
  slug: string;
  title: string;
  summary: string;
  html: string;
  published_at: string;
  metadata: {cover_asset_id?:string;tags?:string[];seo_title?:string;seo_description?:string};
  alternates?:{locale:string;type:string;slug:string}[];
};
const origin = process.env.API_ORIGIN ?? "http://localhost:8080";
const workspace = process.env.WORKSPACE_SLUG ?? "personal";
export type PublicSettings = { workspace:{name:string;default_locale:string;supported_locales:string[]}; site:{name:string;public_url:string;description:string;about:string;footer:string;rss_enabled:boolean;theme?:string;accent_color?:string;share_footer?:string} };
export type PreviewPage = {type:string;locale:string;title:string;summary:string;html:string;expires_at:string};
export async function getPublicSettings():Promise<PublicSettings>{
  const fallback:PublicSettings={workspace:{name:"Personal Workspace",default_locale:"zh-CN",supported_locales:["zh-CN","en"]},site:{name:"Field Notes",public_url:"",description:"",about:"",footer:"Capture · Connect · Publish",rss_enabled:true}};
  try{const response=await fetch(`${origin}/v1/public/${workspace}/settings`,{cache:"no-store"});return response.ok?await response.json():fallback}catch{return fallback}
}
export async function getPage(
  locale: string,
  type: string,
  slug: string,
): Promise<PublishedPage | null> {
  const response = await fetch(
    `${origin}/v1/public/${workspace}/${locale}/${type}/${slug}`,
    { cache: "no-store" },
  );
  if (response.status === 404) return null;
  if (!response.ok) throw new Error("Published content is unavailable");
  return response.json();
}
export async function listPages(locale: string,filters:{q?:string;tag?:string;type?:string}={}): Promise<PublishedPage[]> {
  try {
    const params=new URLSearchParams(Object.entries(filters).filter(([,value])=>value) as [string,string][]);
    const response = await fetch(
      `${origin}/v1/public/${workspace}/${locale}/contents${params.size?`?${params}`:""}`,
      { cache: "no-store" },
    );
    if (!response.ok) return [];
    return (await response.json()).items;
  } catch {
    return [];
  }
}
export type PublicCollection={title:string;slug:string;sections:{title:string;items:PublishedPage[]}[]};
export async function listTags(locale:string):Promise<{name:string;count:number}[]>{try{const response=await fetch(`${origin}/v1/public/${workspace}/${locale}/tags`,{cache:"no-store"});return response.ok?(await response.json()).items:[]}catch{return []}}
export async function listCollections(locale:string):Promise<PublicCollection[]>{try{const response=await fetch(`${origin}/v1/public/${workspace}/${locale}/collections`,{cache:"no-store"});return response.ok?(await response.json()).items:[]}catch{return []}}
export async function getCollection(locale:string,slug:string):Promise<PublicCollection|null>{try{const response=await fetch(`${origin}/v1/public/${workspace}/${locale}/collections/${encodeURIComponent(slug)}`,{cache:"no-store"});if(response.status===404)return null;return response.ok?response.json():null}catch{return null}}
export function publicBase(settings:PublicSettings){return settings.site.public_url||process.env.PUBLIC_SITE_ORIGIN||"http://localhost:3000"}
export async function getPreview(token:string):Promise<PreviewPage|null>{
  try{const response=await fetch(`${origin}/v1/previews/${encodeURIComponent(token)}`,{cache:"no-store"});if(response.status===404)return null;if(!response.ok)throw new Error("Preview is unavailable");return response.json()}catch{return null}
}
