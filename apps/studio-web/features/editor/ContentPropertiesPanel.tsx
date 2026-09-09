import { useEffect, useState } from "react";
import { CheckCircle2, Image as ImageIcon, PanelRightClose, TriangleAlert } from "lucide-react";
import { api, Asset, Content, Draft, Localization, Readiness } from "../../lib/api";

export function ContentPropertiesPanel({workspace,content,localization,draft,canEdit,readiness,onMetadataChange,onSaved,onClose}:{workspace:string;content:Content;localization:Localization;draft:Draft;canEdit:boolean;readiness:Readiness|null;onMetadataChange:(metadata:Record<string,unknown>)=>void;onSaved:()=>void;onClose:()=>void}){
  const [slug,setSlug]=useState(localization.slug),[visibility,setVisibility]=useState(content.visibility),[assets,setAssets]=useState<Asset[]>([]),[status,setStatus]=useState("");
  useEffect(()=>{void api.assets(workspace).then(v=>setAssets(v.items)).catch(()=>setAssets([]))},[workspace]);
  const metadata=draft.metadata??{};
  const set=(key:string,value:unknown)=>onMetadataChange({...metadata,[key]:value});
  async function save(){setStatus("保存中…");try{await api.updateContentProperties(workspace,content.id,{localization_id:localization.id,slug,visibility});onSaved();setStatus("内容属性已保存")}catch(error){setStatus(error instanceof Error?error.message:"保存失败")}}
  return <aside className="properties-panel" aria-label="内容属性">
    <header><div><span className="eyebrow">CONTENT DETAILS</span><h2>内容属性</h2></div><button className="icon-button" aria-label="关闭内容属性" onClick={onClose}><PanelRightClose/></button></header>
    <div className="property-fields">
      <label><span>内容类型</span><input value={content.type} disabled/></label>
      <label><span>固定链接</span><small>{localization.locale==="zh-CN"?"中文":"English"}</small><input disabled={!canEdit} value={slug} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" onChange={e=>setSlug(e.target.value.toLowerCase())}/></label>
      <label><span>可见性</span><select disabled={!canEdit} value={visibility} onChange={e=>setVisibility(e.target.value)}><option value="private">私有</option><option value="unlisted">不公开列出</option><option value="public">公开</option></select></label>
      <label><span>封面图片</span><select disabled={!canEdit} value={String(metadata.cover_asset_id??"")} onChange={e=>set("cover_asset_id",e.target.value)}><option value="">未设置</option>{assets.filter(a=>a.media_type==="image").map(a=><option value={a.id} key={a.id}>{a.filename}</option>)}</select><small><ImageIcon aria-hidden size={14}/> 来自资产库</small></label>
      <label><span>标签</span><input disabled={!canEdit} value={Array.isArray(metadata.tags)?metadata.tags.join(", "):""} onChange={e=>set("tags",e.target.value.split(",").map(v=>v.trim()).filter(Boolean))} placeholder="设计, 随笔, 技术"/></label>
      <label><span>SEO 标题</span><input disabled={!canEdit} value={String(metadata.seo_title??"")} onChange={e=>set("seo_title",e.target.value)}/></label>
      <label><span>SEO 描述</span><textarea disabled={!canEdit} value={String(metadata.seo_description??"")} onChange={e=>set("seo_description",e.target.value)}/></label>
    </div>
    {readiness&&<section className="readiness"><h3>发布前检查</h3>{readiness.issues.length===0?<p className="ready-ok"><CheckCircle2 aria-hidden/>已满足发布条件</p>:<ul>{readiness.issues.map(issue=><li className={issue.severity} key={issue.code}>{issue.severity==="error"?<TriangleAlert aria-hidden/>:<CheckCircle2 aria-hidden/>}<span>{issue.message}<small>{issue.severity==="error"?"发布前必须处理":"建议完善，不阻止发布"}</small></span></li>)}</ul>}</section>}
    {canEdit&&<button className="primary properties-save" onClick={()=>void save()}>保存属性</button>}<p className="settings-status" aria-live="polite">{status}</p>
  </aside>;
}
