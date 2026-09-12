"use client";

import { useEffect, useState } from "react";
import { Check, Copy, KeyRound, ShieldCheck } from "lucide-react";
import { api, TOTPEnrollment } from "../../lib/api";

type View = "status" | "enroll" | "recovery" | "disable";

export function TOTPSettings({canEdit}:{canEdit:boolean}) {
  const [enabled,setEnabled]=useState(false);
  const [loading,setLoading]=useState(true);
  const [busy,setBusy]=useState(false);
  const [view,setView]=useState<View>("status");
  const [enrollment,setEnrollment]=useState<TOTPEnrollment>();
  const [code,setCode]=useState("");
  const [recoveryCodes,setRecoveryCodes]=useState<string[]>([]);
  const [message,setMessage]=useState("");

  useEffect(()=>{
    api.totpStatus()
      .then(value=>setEnabled(value.enabled))
      .catch(err=>setMessage(err instanceof Error?err.message:"无法读取动态码状态"))
      .finally(()=>setLoading(false));
  },[]);

  async function begin(){
    setBusy(true);setMessage("");
    try {setEnrollment(await api.setupTOTP());setCode("");setView("enroll")}
    catch(err){setMessage(err instanceof Error?err.message:"无法开始配置")}
    finally{setBusy(false)}
  }

  async function enable(){
    if(!enrollment)return;
    setBusy(true);setMessage("");
    try {
      const result=await api.enableTOTP(enrollment.secret,code);
      setEnabled(true);setRecoveryCodes(result.recovery_codes);setCode("");setEnrollment(undefined);setView("recovery");
    } catch(err){setMessage(err instanceof Error?err.message:"动态码验证失败，请检查时间后重试")}
    finally{setBusy(false)}
  }

  async function disable(){
    setBusy(true);setMessage("");
    try {await api.disableTOTP(code);setEnabled(false);setCode("");setView("status");setMessage("TOTP 动态码已停用")}
    catch(err){setMessage(err instanceof Error?err.message:"停用失败")}
    finally{setBusy(false)}
  }

  async function copy(value:string,success:string){
    await navigator.clipboard.writeText(value);
    setMessage(success);
  }

  return <section className="settings-card settings-card-wide totp-settings" id="settings-security">
    <header>
      <div><h2>登录安全</h2><p>绑定验证器后，可只使用邮箱和动态码登录，不需要密码。</p></div>
      <span className={`totp-badge ${enabled?"enabled":""}`}><ShieldCheck aria-hidden size={16}/>{loading?"读取中":enabled?"已绑定":"未绑定"}</span>
    </header>

    {view==="status"&&<div className="totp-summary">
      <div className="totp-summary-icon"><KeyRound aria-hidden/></div>
      <div><strong>{enabled?"动态码登录已就绪":"添加 TOTP 验证器"}</strong><p>{enabled?"验证器每 30 秒生成一个 6 位动态码。登录开关由下方“登录与发布”设置控制。":"支持常见验证器应用，绑定后仍可继续使用密码或 Passkey。"}</p></div>
      {canEdit&&(enabled?<button className="danger" onClick={()=>{setCode("");setMessage("");setView("disable")}}>停用</button>:<button className="primary compact" disabled={loading||busy} onClick={()=>void begin()}>{busy?"准备中…":"开始绑定"}</button>)}
    </div>}

    {view==="enroll"&&enrollment&&<div className="totp-enrollment">
      <div className="totp-qr"><img src={enrollment.qr_data_url} width={200} height={200} alt="TOTP 验证器绑定二维码"/></div>
      <div className="totp-steps">
        <span className="eyebrow">TWO QUICK STEPS</span>
        <h3>扫描二维码并验证</h3>
        <ol><li>使用验证器应用扫描二维码。</li><li>输入当前显示的 6 位动态码完成绑定。</li></ol>
        <label className="totp-secret"><span>无法扫码？手动输入密钥</span><code>{enrollment.secret}</code><button className="secondary compact" type="button" onClick={()=>void copy(enrollment.secret,"密钥已复制")}><Copy aria-hidden size={15}/>复制</button></label>
        <label className="settings-field"><span>6 位动态码</span><input value={code} onChange={e=>setCode(e.target.value.replace(/\D/g,"").slice(0,6))} inputMode="numeric" autoComplete="one-time-code" placeholder="000000" autoFocus/></label>
        <div className="totp-actions"><button className="secondary compact" onClick={()=>{setEnrollment(undefined);setCode("");setView("status")}}>取消</button><button className="primary compact" disabled={busy||code.length!==6} onClick={()=>void enable()}>{busy?"验证中…":"验证并启用"}</button></div>
      </div>
    </div>}

    {view==="recovery"&&<div className="totp-recovery">
      <div className="totp-success"><Check aria-hidden/><div><h3>动态码已启用</h3><p>请现在保存以下 8 个恢复码。每个只能使用一次，离开此页后不再显示。</p></div></div>
      <div className="recovery-code-grid">{recoveryCodes.map(value=><code key={value}>{value}</code>)}</div>
      <div className="totp-actions"><button className="secondary compact" onClick={()=>void copy(recoveryCodes.join("\n"),"恢复码已复制")}><Copy aria-hidden size={15}/>复制全部</button><button className="primary compact" onClick={()=>{setRecoveryCodes([]);setMessage("");setView("status")}}>我已保存</button></div>
    </div>}

    {view==="disable"&&<div className="totp-disable">
      <h3>停用 TOTP 动态码</h3><p>输入当前 6 位动态码或一个未使用的恢复码进行确认。</p>
      <label className="settings-field"><span>动态码或恢复码</span><input value={code} onChange={e=>setCode(e.target.value.toUpperCase())} autoComplete="one-time-code" autoFocus/></label>
      <div className="totp-actions"><button className="secondary compact" onClick={()=>{setCode("");setMessage("");setView("status")}}>取消</button><button className="danger" disabled={busy||!code.trim()} onClick={()=>void disable()}>{busy?"停用中…":"确认停用"}</button></div>
    </div>}

    {message&&<p className="settings-status" role="status">{message}</p>}
  </section>
}
