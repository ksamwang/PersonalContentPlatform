"use client";
import { useEffect, useState } from "react";
import { api, AISettings, GeneralSettings, StorageSettings } from "../../lib/api";
import { GeneralSettingsForm } from "./GeneralSettingsForm";
import { ProviderSettingsForm } from "./ProviderSettingsForm";
import { WebhookSettings } from "./WebhookSettings";

export function SettingsPanel({workspace,role}:{workspace:string;role:string}){
 const [general,setGeneral]=useState<GeneralSettings>(),[storage,setStorage]=useState<StorageSettings>(),[ai,setAI]=useState<AISettings>(),[error,setError]=useState("");
 useEffect(()=>{api.settings(workspace).then(v=>{setGeneral(v.general);setStorage(v.storage??{name:"Default",provider:"filesystem",endpoint:"",region:"",bucket:"",base_path:"./data/objects"});setAI(v.ai)}).catch(e=>setError(e instanceof Error?e.message:"设置加载失败"))},[workspace]);
 if(error)return <section className="settings-page" role="alert"><h1>设置加载失败</h1><p>{error}</p></section>;
 if(!general||!storage||!ai)return <section className="settings-page" aria-busy="true"><h1>工作区设置</h1><p>正在加载…</p></section>;
 const canEdit=role==="owner";
 return <div className="settings-page"><header className="settings-heading"><span className="eyebrow">WORKSPACE CONTROL</span><h1>工作区设置</h1><p>{canEdit?"配置内容站点、Provider 与发布行为。":"你可以查看配置，只有 Owner 可以修改。"}</p></header><GeneralSettingsForm value={general} setValue={setGeneral} canEdit={canEdit} save={async section=>{const saved=await api.saveSettings(workspace,section,general[section]);setGeneral(saved)}}/><ProviderSettingsForm workspace={workspace} storage={storage} setStorage={setStorage} ai={ai} setAI={setAI} canEdit={canEdit}/><WebhookSettings workspace={workspace} canEdit={canEdit}/></div>
}
