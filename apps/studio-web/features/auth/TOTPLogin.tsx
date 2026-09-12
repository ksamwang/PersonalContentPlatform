"use client";

import { FormEvent, useState } from "react";
import { api, Principal } from "../../lib/api";

export function TOTPLogin({email,setEmail,onAuthenticated}:{email:string;setEmail:(value:string)=>void;onAuthenticated:(p:Principal)=>void}) {
  const [code,setCode]=useState("");
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState("");

  async function submit(event:FormEvent<HTMLFormElement>){
    event.preventDefault();setBusy(true);setError("");
    try {onAuthenticated(await api.loginTOTP(email,code))}
    catch(err){setError(err instanceof Error?err.message:"无法使用动态码登录")}
    finally{setBusy(false)}
  }

  return <form onSubmit={submit} className="totp-login-form">
    <label>邮箱<input type="email" autoComplete="email" value={email} onChange={event=>setEmail(event.target.value)} required autoFocus/></label>
    <label>动态码或恢复码<input value={code} onChange={event=>setCode(event.target.value.toUpperCase().replace(/\s/g,""))} inputMode="text" autoComplete="one-time-code" placeholder="6 位动态码" required/><small>也可以输入绑定时保存的一次性恢复码</small></label>
    {error&&<p className="form-error" role="alert">{error}</p>}
    <button className="primary" disabled={busy||!email||!code}>{busy?"正在验证…":"使用动态码进入"}</button>
  </form>
}
