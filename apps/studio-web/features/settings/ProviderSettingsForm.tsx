import { useState } from "react";
import { api, AISettings, StorageSettings } from "../../lib/api";
import { Field, SettingsSection } from "./SettingsSection";

export function ProviderSettingsForm({workspace,storage,setStorage,ai,setAI,canEdit}:{workspace:string;storage:StorageSettings;setStorage:(v:StorageSettings)=>void;ai:AISettings;setAI:(v:AISettings)=>void;canEdit:boolean}){
 const [accessKey,setAccessKey]=useState(""),[secretKey,setSecretKey]=useState(""),[apiKey,setAPIKey]=useState(""),[storageProbe,setStorageProbe]=useState(""),[aiProbe,setAIProbe]=useState("");
 const setS=(key:keyof StorageSettings,value:unknown)=>setStorage({...storage,[key]:value});
 return <>
  <SettingsSection title="对象存储" description="新上传使用当前配置；旧文件继续绑定原有 Provider。" canEdit={canEdit} onSave={async()=>{const saved=await api.saveStorage(workspace,{...storage,access_key:accessKey,secret_key:secretKey});setStorage(saved);setAccessKey("");setSecretKey("")}}>
   <Field label="Provider"><select disabled={!canEdit} value={storage.provider} onChange={e=>setS("provider",e.target.value)}><option value="filesystem">本地文件系统</option><option value="oss">阿里云 OSS</option><option value="s3">S3</option><option value="r2">Cloudflare R2</option></select></Field>
   {storage.provider==="filesystem"?<Field label="存储子目录" hint="只能使用部署时配置的存储根目录或其子目录。"><input disabled={!canEdit} value={storage.base_path} onChange={e=>setS("base_path",e.target.value)}/></Field>:<>
    <Field label="Endpoint"><input disabled={!canEdit} value={storage.endpoint} onChange={e=>setS("endpoint",e.target.value)}/></Field><Field label="Region"><input disabled={!canEdit} value={storage.region} onChange={e=>setS("region",e.target.value)}/></Field><Field label="Bucket"><input disabled={!canEdit} value={storage.bucket} onChange={e=>setS("bucket",e.target.value)}/></Field>
    <Field label="Access Key" hint={storage.access_key_mask?`当前：${storage.access_key_mask}；留空保持不变`:undefined}><input disabled={!canEdit} autoComplete="off" value={accessKey} onChange={e=>setAccessKey(e.target.value)}/></Field><Field label="Secret Key" hint={storage.secret_key_set?"已配置；留空保持不变":undefined}><input disabled={!canEdit} type="password" autoComplete="new-password" value={secretKey} onChange={e=>setSecretKey(e.target.value)}/></Field>
   </>}
   {canEdit&&<button className="secondary settings-test" onClick={async()=>{setStorageProbe("测试中…");try{await api.testStorage(workspace);setStorageProbe("连接成功")}catch(e){setStorageProbe(e instanceof Error?e.message:"连接失败")}}}>测试连接</button>}<span aria-live="polite">{storageProbe}</span>
  </SettingsSection>
  <SettingsSection title="AI Provider" description="支持 OpenAI 兼容接口，可为不同用途指定模型。" canEdit={canEdit} onSave={async()=>{const saved=await api.saveAI(workspace,{...ai,api_key:apiKey});setAI(saved);setAPIKey("")}}>
   <Field label="接口类型"><select disabled={!canEdit} value={ai.provider} onChange={e=>setAI({...ai,provider:e.target.value})}><option value="openai-compatible">OpenAI Compatible</option></select></Field><Field label="Base URL"><input disabled={!canEdit} type="url" value={ai.base_url} onChange={e=>setAI({...ai,base_url:e.target.value})}/></Field><Field label="默认模型"><input disabled={!canEdit} value={ai.model} onChange={e=>setAI({...ai,model:e.target.value})}/></Field><Field label="API Key" hint={ai.api_key_set?`当前：${ai.api_key_mask}；留空保持不变`:undefined}><input disabled={!canEdit} type="password" autoComplete="new-password" value={apiKey} onChange={e=>setAPIKey(e.target.value)}/></Field>
   {canEdit&&<button className="secondary settings-test" onClick={async()=>{setAIProbe("测试中…");try{await api.testAI(workspace);setAIProbe("连接成功")}catch(e){setAIProbe(e instanceof Error?e.message:"连接失败")}}}>测试连接</button>}<span aria-live="polite">{aiProbe}</span>
  </SettingsSection>
 </>
}
