import { FormEvent, useState } from "react";
import { Languages, Plus, Sparkles, X } from "lucide-react";
import { api, Content } from "../../lib/api";

const names:Record<string,string>={"zh-CN":"中文","en":"English"};
const stateNames:Record<string,string>={missing:"待翻译",draft:"草稿",needs_review:"待校对",ready:"已就绪",published:"已发布",outdated:"需要同步"};

export function LocalizationTabs({workspace,content,active,canEdit,translating,onSelect,onCreated,onTranslate}:{workspace:string;content:Content;active:string;canEdit:boolean;translating:boolean;onSelect:(id:string)=>void;onCreated:(content:Content,targetID:string)=>void;onTranslate:()=>void}){
  const [adding,setAdding]=useState(false),[slug,setSlug]=useState(""),[error,setError]=useState(""),[busy,setBusy]=useState(false);
  const missing=["zh-CN","en"].find(locale=>!content.localizations.some(item=>item.locale===locale));
  const source=content.localizations.find(item=>item.id===active)??content.localizations[0];
  const isTranslation=source.locale!==content.default_locale&&content.localizations.length>1;
  function begin(){if(!missing)return;setSlug(`${source.slug}-${missing==="en"?"en":"zh"}`);setAdding(true);setError("")}
  async function create(event:FormEvent){event.preventDefault();if(!missing)return;setBusy(true);try{const next=await api.createLocalization(workspace,content.id,{locale:missing,slug,source_locale:source.locale});const target=next.localizations.find(item=>item.locale===missing);if(target)onCreated(next,target.id);setAdding(false)}catch(value){setError(value instanceof Error?value.message:"语言版本创建失败")}finally{setBusy(false)}}
  return <div className="localization-bar"><div className="localization-tabs" role="tablist" aria-label="内容语言">{content.localizations.map(item=><button type="button" role="tab" aria-selected={item.id===active} key={item.id} onClick={()=>onSelect(item.id)}><span>{names[item.locale]??item.locale}</span><small>{stateNames[item.translation_status]??item.state}</small></button>)}{canEdit&&missing&&!adding&&<button type="button" className="add-localization" onClick={begin}><Plus aria-hidden/>添加 {names[missing]}</button>}{canEdit&&isTranslation&&!adding&&<button type="button" className="translate-action" disabled={translating} onClick={onTranslate}><Sparkles aria-hidden/>{translating?"AI 翻译中…":source.translation_status==="outdated"?"AI 同步译稿":"AI 生成译稿"}</button>}</div>{adding&&<form className="localization-create" onSubmit={create}><Languages aria-hidden/><label><span className="sr-only">新语言固定链接</span><input required pattern="[a-z0-9]+(?:-[a-z0-9]+)*" value={slug} onChange={e=>setSlug(e.target.value.toLowerCase())}/></label><button className="primary compact" disabled={busy}>{busy?"创建中":"创建语言版本"}</button><button type="button" className="icon-button" aria-label="取消" onClick={()=>setAdding(false)}><X/></button>{error&&<span role="alert">{error}</span>}</form>}</div>;
}
