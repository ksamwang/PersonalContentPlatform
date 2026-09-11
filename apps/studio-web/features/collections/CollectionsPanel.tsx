import { useCallback, useEffect, useState } from "react";
import { Plus, Shapes } from "lucide-react";
import { api, Collection } from "../../lib/api";
import { CollectionDetail } from "./CollectionDetail";
import { CreateCollectionDialog } from "./CreateCollectionDialog";
import { WorkspacePage } from "../studio/WorkspacePage";

export function CollectionsPanel({workspace,role,onOpenContent}:{workspace:string;role:string;onOpenContent:(id:string)=>void}){
 const [items,setItems]=useState<Collection[]>([]),[selected,setSelected]=useState<Collection>(),[open,setOpen]=useState(false),[loading,setLoading]=useState(true),[error,setError]=useState("");const canEdit=role==="owner"||role==="editor";
 const loadList=useCallback(async()=>{setLoading(true);setError("");try{const result=await api.collections(workspace);setItems(result.items);if(!selected&&result.items[0])setSelected(await api.collection(workspace,result.items[0].id))}catch(err){setError(err instanceof Error?err.message:"合集加载失败")}finally{setLoading(false)}},[workspace,selected]);
 useEffect(()=>{void loadList()},[workspace]);
 async function select(id:string){try{setSelected(await api.collection(workspace,id));setError("")}catch(err){setError(err instanceof Error?err.message:"合集详情加载失败")}}
 async function reload(){if(selected)setSelected(await api.collection(workspace,selected.id));const result=await api.collections(workspace);setItems(result.items)}
 return <WorkspacePage className="collections-page" eyebrow="CURATED SETS" title="合集" description="按主题、路径或展示顺序，把已有内容编排成可维护的集合。" actions={canEdit&&<button className="primary compact" onClick={()=>setOpen(true)}><Plus aria-hidden size={18}/>新建合集</button>}>
  {error&&<div className="inline-error" role="alert"><span>{error}</span><button onClick={()=>void loadList()}>重试</button></div>}
  <div className="collections-grid"><aside className="collection-index" aria-busy={loading}>{loading?<p>正在加载…</p>:items.length===0?<div className="collection-empty"><Shapes aria-hidden/><h2>还没有合集</h2><p>创建第一个合集来组织内容。</p></div>:items.map(item=><button key={item.id} className={selected?.id===item.id?"active":""} onClick={()=>void select(item.id)}><strong>{item.title}</strong><small>/{item.slug}</small></button>)}</aside>{selected?<CollectionDetail workspace={workspace} value={selected} canEdit={canEdit} onReload={reload} onOpenContent={onOpenContent}/>:<div className="collection-empty"><h2>选择一个合集</h2><p>详情、分区和内容顺序会显示在这里。</p></div>}</div>
  {open&&<CreateCollectionDialog workspace={workspace} onClose={()=>setOpen(false)} onCreated={value=>{setItems(v=>[value,...v]);setSelected(value);setOpen(false)}}/>}
 </WorkspacePage>
}
