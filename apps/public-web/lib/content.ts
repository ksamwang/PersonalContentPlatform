export type PublishedPage = {
  locale: string;
  type: string;
  slug: string;
  title: string;
  summary: string;
  html: string;
  published_at: string;
};
const origin = process.env.API_ORIGIN ?? "http://localhost:8080";
const workspace = process.env.WORKSPACE_SLUG ?? "personal";
export type PublicSettings = { workspace:{name:string;default_locale:string;supported_locales:string[]}; site:{name:string;public_url:string;description:string;about:string;footer:string;rss_enabled:boolean} };
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
export async function listPages(locale: string): Promise<PublishedPage[]> {
  try {
    const response = await fetch(
      `${origin}/v1/public/${workspace}/${locale}/contents`,
      { cache: "no-store" },
    );
    if (!response.ok) return [];
    return (await response.json()).items;
  } catch {
    return [];
  }
}
