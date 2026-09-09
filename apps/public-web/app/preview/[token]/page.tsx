import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getPreview } from "../../../lib/content";

export const metadata:Metadata={title:"草稿预览",robots:{index:false,follow:false}};
export default async function Preview({params}:{params:Promise<{token:string}>}){
 const {token}=await params,page=await getPreview(token);if(!page)notFound();
 return <><div className="preview-banner" role="status"><strong>草稿预览</strong><span>此页面不会公开收录，链接有效至 {new Intl.DateTimeFormat("zh-CN",{dateStyle:"medium",timeStyle:"short"}).format(new Date(page.expires_at))}</span></div><main id="content" className="article-shell preview-shell"><article><header><span className="kicker">{page.type} · {page.locale}</span><h1>{page.title}</h1>{page.summary&&<p className="dek">{page.summary}</p>}</header><div className="article-body" dangerouslySetInnerHTML={{__html:page.html}}/></article></main></>
}
