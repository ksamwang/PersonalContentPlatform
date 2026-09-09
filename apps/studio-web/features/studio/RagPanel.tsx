import { FormEvent, useState } from "react";
import { Bot, Send } from "lucide-react";
import { api, SearchHit } from "../../lib/api";

export function RagPanel({workspace,locale,onOpen}:{workspace:string;locale:string;onOpen:(id:string)=>void}) {
  const [question,setQuestion]=useState(""),[answer,setAnswer]=useState(""),[sources,setSources]=useState<SearchHit[]>([]),[status,setStatus]=useState("");
  async function ask(event:FormEvent){event.preventDefault();if(!question.trim())return;setStatus("正在查找资料并生成回答…");setAnswer("");try{const result=await api.rag(workspace,question.trim(),locale);setAnswer(result.answer);setSources(result.sources??[]);setStatus("")}catch(error){setStatus(error instanceof Error?error.message:"问答失败")}}
  return <section className="rag-panel">
    <div className="rag-heading"><Bot aria-hidden/><div><h2>知识问答</h2><p>回答只使用你的内容，并标注来源编号。</p></div></div>
    <form onSubmit={ask}><textarea value={question} onChange={e=>setQuestion(e.target.value)} placeholder="例如：我对这个主题写过哪些观点？"/><button className="primary"><Send aria-hidden size={16}/>提问</button></form>
    {status&&<p className="inline-status" role="status">{status}</p>}
    {answer&&<div className="rag-answer"><p>{answer}</p>{sources.length>0&&<ol>{sources.map((source,index)=><li key={`${source.content_id}-${index}`}><button onClick={()=>onOpen(source.content_id)}>[{index+1}] {source.title}</button><span>{source.locale}</span></li>)}</ol>}</div>}
  </section>;
}
