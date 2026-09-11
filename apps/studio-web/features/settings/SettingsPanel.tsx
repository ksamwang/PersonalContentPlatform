"use client";
import { useEffect, useState } from "react";
import { api, AISettings, EmbeddingSettings, GeneralSettings, MediaSettings, StorageSettings } from "../../lib/api";
import { GeneralSettingsForm } from "./GeneralSettingsForm";
import { ProviderSettingsForm } from "./ProviderSettingsForm";
import { WebhookSettings } from "./WebhookSettings";
import { EmbeddingSettingsForm } from "./EmbeddingSettingsForm";
import { MediaProviderSettings } from "./MediaProviderSettings";
import { WorkspacePage } from "../studio/WorkspacePage";

export function SettingsPanel({workspace,role}:{workspace:string;role:string}){
 const [general,setGeneral]=useState<GeneralSettings>(),[storage,setStorage]=useState<StorageSettings>(),[ai,setAI]=useState<AISettings>(),[embedding,setEmbedding]=useState<EmbeddingSettings>(),[media,setMedia]=useState<MediaSettings[]>(),[error,setError]=useState("");
 useEffect(()=>{api.settings(workspace).then(v=>{setGeneral(v.general);setStorage(v.storage??{name:"Default",provider:"filesystem",endpoint:"",region:"",bucket:"",base_path:"./data/objects"});setAI(v.ai);setEmbedding(v.embedding??{provider:"openai-compatible",base_url:"",model:"",dimensions:0});setMedia(v.media?.length?v.media:[{purpose:"ocr",provider:"openai-compatible",base_url:"",model:""},{purpose:"transcription",provider:"openai-compatible",base_url:"",model:""}])}).catch(e=>setError(e instanceof Error?e.message:"设置加载失败"))},[workspace]);
 if(error)return <WorkspacePage className="settings-page" eyebrow="WORKSPACE CONTROL" title="设置加载失败" description={error}><div role="alert" className="settings-loading">请检查服务连接后刷新页面。</div></WorkspacePage>;
 if(!general||!storage||!ai||!embedding||!media)return <WorkspacePage className="settings-page" eyebrow="WORKSPACE CONTROL" title="工作区设置" description="配置内容站点、Provider 与发布行为。"><div className="settings-loading" aria-busy="true">正在加载设置…</div></WorkspacePage>;
 const canEdit=role==="owner";
 return <WorkspacePage className="settings-page" eyebrow="WORKSPACE CONTROL" title="工作区设置" description={canEdit?"配置内容站点、Provider 与发布行为。":"你可以查看配置，只有 Owner 可以修改。"}>
  <div className="settings-layout">
   <nav className="settings-nav" aria-label="设置目录">
    <a href="#settings-workspace">工作区与站点</a>
    <a href="#settings-storage">对象存储</a>
    <a href="#settings-ai">AI 创作</a>
    <a href="#settings-embedding">语义检索</a>
    <a href="#settings-media">媒体处理</a>
    <a href="#settings-webhook">Webhook</a>
   </nav>
   <div className="settings-content">
    <GeneralSettingsForm value={general} setValue={setGeneral} canEdit={canEdit} save={async section=>{const saved=await api.saveSettings(workspace,section,general[section]);setGeneral(saved)}}/>
    <ProviderSettingsForm workspace={workspace} storage={storage} setStorage={setStorage} ai={ai} setAI={setAI} canEdit={canEdit}/>
    <div id="settings-embedding"><EmbeddingSettingsForm workspace={workspace} value={embedding} setValue={setEmbedding} canEdit={canEdit}/></div>
    <MediaProviderSettings workspace={workspace} values={media} setValues={setMedia} canEdit={canEdit}/>
    <WebhookSettings workspace={workspace} canEdit={canEdit}/>
   </div>
  </div>
 </WorkspacePage>
}
