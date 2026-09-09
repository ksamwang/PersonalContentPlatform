import { useCallback, useEffect, useState } from "react";
import { RefreshCw } from "lucide-react";
import { api, AISettings } from "../../lib/api";
import { Field, SettingsSection } from "./SettingsSection";

export function AIProviderSettings({workspace,ai,setAI,canEdit}:{workspace:string;ai:AISettings;setAI:(value:AISettings)=>void;canEdit:boolean}){
  const [apiKey,setAPIKey]=useState(""),[models,setModels]=useState<string[]>([]),[modelState,setModelState]=useState<"idle"|"loading"|"ready"|"error">("idle"),[modelMessage,setModelMessage]=useState(""),[probe,setProbe]=useState("");
  const loadModels=useCallback(async(key=apiKey)=>{
    if(!ai.base_url){setModelState("error");setModelMessage("请先填写 Base URL。 ");return}
    setModelState("loading");setModelMessage("正在从 Provider 获取模型…");
    try{const result=await api.aiModels(workspace,{provider:ai.provider,base_url:ai.base_url,api_key:key||undefined});const ids=result.items.map(item=>item.id);setModels(ids);setModelState("ready");setModelMessage(`已获取 ${ids.length} 个模型。`);if(!ids.includes(ai.model))setAI({...ai,model:ids[0]??""})}
    catch(error){setModels(ai.model?[ai.model]:[]);setModelState("error");setModelMessage(error instanceof Error?`${error.message}；请检查地址和 API Key 后重试。`:"模型获取失败，请重试。")}
  },[workspace,ai,apiKey,setAI]);

  useEffect(()=>{if(ai.base_url&&ai.api_key_set&&models.length===0&&modelState==="idle")void loadModels("")},[ai.base_url,ai.api_key_set,models.length,modelState,loadModels]);

  function updateBaseURL(value:string){setAI({...ai,base_url:value});setModels([]);setModelState("idle");setModelMessage("修改连接信息后，请重新获取模型。")}
  function updateAPIKey(value:string){setAPIKey(value);setModels([]);setModelState("idle");setModelMessage("输入新密钥后，请获取模型。")}
  return <SettingsSection title="AI Provider" description="支持 OpenAI / Anthropic 兼容接口；模型列表由 Provider 端点提供。" canEdit={canEdit} onSave={async()=>{
    if(!ai.model)throw new Error("请先获取并选择默认模型")
    const saved=await api.saveAI(workspace,{provider:ai.provider,base_url:ai.base_url,model:ai.model,purpose_models:ai.purpose_models,api_key:apiKey||undefined});setAI(saved);setAPIKey("");setProbe("")
  }}>
    <Field label="接口类型"><select disabled={!canEdit} value={ai.provider} onChange={event=>{setAI({...ai,provider:event.target.value});setModels([]);setModelState("idle");setModelMessage("切换协议后，请重新获取模型。")}}><option value="openai-compatible">OpenAI Compatible</option><option value="anthropic-compatible">Anthropic Compatible</option></select></Field>
    <Field label="Base URL" hint="填写兼容 API 的基础地址，例如 https://api.deepseek.com 或包含 /v1 的服务地址。"><input disabled={!canEdit} type="url" value={ai.base_url} onChange={event=>updateBaseURL(event.target.value)}/></Field>
    <Field label="API Key" hint={ai.api_key_set?`当前：${ai.api_key_mask}；留空使用已保存密钥`:"模型发现和调用均需要 API Key。"}><input disabled={!canEdit} type="password" autoComplete="new-password" value={apiKey} onChange={event=>updateAPIKey(event.target.value)}/></Field>
    <Field label="默认模型" hint={modelMessage||"先获取模型，再选择默认模型。"}><div className="model-select-row"><select disabled={!canEdit||modelState==="loading"||models.length===0} value={ai.model} onChange={event=>setAI({...ai,model:event.target.value})}><option value="">{modelState==="loading"?"正在获取模型…":"请选择模型"}</option>{models.map(model=><option key={model} value={model}>{model}</option>)}</select>{canEdit&&<button type="button" className="secondary compact model-refresh" disabled={modelState==="loading"||!ai.base_url} onClick={()=>void loadModels()}><RefreshCw aria-hidden size={16}/>{modelState==="loading"?"获取中…":models.length?"刷新模型":"获取模型"}</button>}</div></Field>
    {canEdit&&<button type="button" className="secondary settings-test" onClick={async()=>{setProbe("测试中…");try{await api.testAI(workspace);setProbe("连接成功")}catch(error){setProbe(error instanceof Error?`${error.message}；请先保存完整配置。`:"连接失败")}}}>测试连接</button>}<span aria-live="polite">{probe}</span>
  </SettingsSection>
}
