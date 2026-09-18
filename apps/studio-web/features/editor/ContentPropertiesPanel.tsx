import { useEffect, useState } from "react";
import { CheckCircle2, Image as ImageIcon, PanelRightClose, TriangleAlert } from "lucide-react";
import { api, Asset, Content, Draft, Localization, Readiness } from "../../lib/api";
import { AssetImage } from "./AssetImage";
import { CoverAssetPickerDialog } from "./CoverAssetPickerDialog";

export function ContentPropertiesPanel({workspace,content,localization,draft,canEdit,readiness,onMetadataChange,onSaved,onClose}:{workspace:string;content:Content;localization:Localization;draft:Draft;canEdit:boolean;readiness:Readiness|null;onMetadataChange:(metadata:Record<string,unknown>)=>void;onSaved:()=>void;onClose:()=>void}){
  const [slug,setSlug]=useState(localization.slug),[visibility,setVisibility]=useState(content.visibility),[assets,setAssets]=useState<Asset[]>([]),[status,setStatus]=useState(""),[coverPickerOpen,setCoverPickerOpen]=useState(false);
  useEffect(()=>{void api.assets(workspace).then(v=>setAssets(v.items)).catch(()=>setAssets([]))},[workspace]);
  const metadata=draft.metadata??{};
  const selectedCover=assets.find(asset=>asset.id===String(metadata.cover_asset_id??""));
  const set=(key:string,value:unknown)=>onMetadataChange({...metadata,[key]:value});
  async function save(){setStatus("保存中…");try{await api.updateContentProperties(workspace,content.id,{localization_id:localization.id,slug,visibility});onSaved();setStatus("内容属性已保存")}catch(error){setStatus(error instanceof Error?error.message:"保存失败")}}
  return <aside className="properties-panel" aria-label="内容属性">
    <header><div><span className="eyebrow">CONTENT DETAILS</span><h2>内容属性</h2></div><button className="icon-button" aria-label="关闭内容属性" onClick={onClose}><PanelRightClose/></button></header>
    <div className="property-fields">
      <label><span>内容类型</span><input value={content.type} disabled/></label>
      <label><span>固定链接</span><small>{localization.locale==="zh-CN"?"中文":"English"}</small><input disabled={!canEdit} value={slug} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" onChange={e=>setSlug(e.target.value.toLowerCase())}/></label>
      <label><span>可见性</span><select disabled={!canEdit} value={visibility} onChange={e=>setVisibility(e.target.value)}><option value="private">私有</option><option value="unlisted">不公开列出</option><option value="public">公开</option></select></label>
      <fieldset className="property-cover-field"><legend>封面图片</legend>{selectedCover?<div className="property-cover-preview"><AssetImage asset={selectedCover} recipe="content-1280" alt={selectedCover.filename}/><span><strong title={selectedCover.filename}>{selectedCover.filename}</strong><small>{selectedCover.width&&selectedCover.height?`${selectedCover.width} × ${selectedCover.height} · `:""}{Math.ceil(selectedCover.size/1024)} KB</small></span>{canEdit&&<div className="property-cover-actions"><button type="button" className="secondary" onClick={()=>setCoverPickerOpen(true)}>更换图片</button><button type="button" className="text-button" onClick={()=>set("cover_asset_id","")}>移除</button></div>}</div>:<button type="button" className="property-cover-empty" disabled={!canEdit} onClick={()=>setCoverPickerOpen(true)}><ImageIcon aria-hidden/><strong>选择封面图片</strong><span>打开图片库查看缩略图后选择</span></button>}</fieldset>
      <label><span>标签</span><input disabled={!canEdit} value={Array.isArray(metadata.tags)?metadata.tags.join(", "):""} onChange={e=>set("tags",e.target.value.split(/[,，、;；\n]+/).map(v=>v.trim()).filter(Boolean))} placeholder="设计, 随笔, 技术"/><small>可使用逗号、顿号或分号分隔</small></label>
      <label><span>SEO 标题</span><input disabled={!canEdit} value={String(metadata.seo_title??"")} onChange={e=>set("seo_title",e.target.value)}/></label>
      <label><span>SEO 描述</span><textarea disabled={!canEdit} value={String(metadata.seo_description??"")} onChange={e=>set("seo_description",e.target.value)}/></label>
    </div>
    {readiness&&<section className="readiness"><h3>发布前检查</h3>{readiness.issues.length===0?<p className="ready-ok"><CheckCircle2 aria-hidden/>已满足发布条件</p>:<ul>{readiness.issues.map(issue=><li className={issue.severity} key={issue.code}>{issue.severity==="error"?<TriangleAlert aria-hidden/>:<CheckCircle2 aria-hidden/>}<span>{issue.message}<small>{issue.severity==="error"?"发布前必须处理":"建议完善，不阻止发布"}</small></span></li>)}</ul>}</section>}
    {canEdit&&<button className="primary properties-save" onClick={()=>void save()}>保存属性</button>}<p className="settings-status" aria-live="polite">{status}</p>
    {coverPickerOpen&&<CoverAssetPickerDialog items={assets.filter(asset=>asset.media_type==="image")} selected={String(metadata.cover_asset_id??"")} onSelect={id=>set("cover_asset_id",id)} onClose={()=>setCoverPickerOpen(false)}/>}
  </aside>;
}
