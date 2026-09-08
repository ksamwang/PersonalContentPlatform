import { FormEvent, useEffect, useState } from "react";
import { api, WebhookEndpoint } from "../../lib/api";
import { Field } from "./SettingsSection";

export function WebhookSettings({workspace,canEdit}:{workspace:string;canEdit:boolean}){
 const [items,setItems]=useState<WebhookEndpoint[]>([]),[status,setStatus]=useState("");
 const load=()=>api.webhooks(workspace).then(v=>setItems(v.items));
 useEffect(()=>{void load().catch(e=>setStatus(e instanceof Error?e.message:"加载失败"))},[workspace]);
 async function create(event:FormEvent<HTMLFormElement>){event.preventDefault();const form=new FormData(event.currentTarget);setStatus("创建中…");try{await api.createWebhook(workspace,{name:String(form.get("name")),url:String(form.get("url")),secret:String(form.get("secret")),event_types:String(form.get("events")).split(",").map(v=>v.trim()).filter(Boolean)});event.currentTarget.reset();await load();setStatus("Webhook 已创建")}catch(e){setStatus(e instanceof Error?e.message:"创建失败")}}
 return <section className="settings-card"><header><div><h2>Webhook</h2><p>管理事件通知地址和签名密钥。密钥保存后不会再次显示。</p></div></header>
  {canEdit&&<form className="settings-fields" onSubmit={e=>void create(e)}><Field label="名称"><input name="name" required/></Field><Field label="URL"><input name="url" type="url" required/></Field><Field label="签名密钥"><input name="secret" type="password" autoComplete="new-password" required/></Field><Field label="事件类型" hint="英文逗号分隔；留空表示全部事件。"><input name="events"/></Field><button className="primary compact settings-test">添加 Webhook</button></form>}
  <div className="webhook-list">{items.length===0?<p>尚未配置 Webhook。</p>:items.map(item=><article key={item.id}><div><strong>{item.name}</strong><span>{item.url}</span><small>{item.secret_set?"签名密钥已配置":"未配置签名密钥"}</small></div><div>{canEdit&&<><button className="secondary" onClick={async()=>{await api.setWebhookEnabled(workspace,item.id,!item.enabled);await load()}}>{item.enabled?"停用":"启用"}</button><button className="danger" onClick={async()=>{if(!window.confirm(`确认删除 ${item.name}？`))return;await api.deleteWebhook(workspace,item.id);await load()}}>删除</button></>}</div></article>)}</div><p className="settings-status" aria-live="polite">{status}</p>
 </section>
}
