import { ReactNode, useState } from "react";

export function SettingsSection({ id, className="", title, description, children, onSave, canEdit }: { id?:string; className?:string; title:string; description:string; children:ReactNode; onSave:()=>Promise<void>; canEdit:boolean }) {
  const [state,setState]=useState<"idle"|"saving"|"saved"|"error">("idle");
  const [error,setError]=useState("");
  async function save(){setState("saving");setError("");try{await onSave();setState("saved")}catch(err){setError(err instanceof Error?err.message:"保存失败");setState("error")}}
  return <section className={`settings-card ${className}`.trim()} id={id}>
    <header><div><h2>{title}</h2><p>{description}</p></div>{canEdit&&<button className="primary compact" disabled={state==="saving"} onClick={()=>void save()}>{state==="saving"?"保存中…":"保存"}</button>}</header>
    <div className="settings-fields">{children}</div>
    <div className={`settings-status ${state}`} aria-live="polite">{state==="saved"?"设置已保存":error}</div>
  </section>
}

export function Field({label,hint,children}:{label:string;hint?:string;children:ReactNode}){return <label className="settings-field"><span>{label}</span>{children}{hint&&<small>{hint}</small>}</label>}
