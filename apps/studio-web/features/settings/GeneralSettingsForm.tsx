import { GeneralSettings } from "../../lib/api";
import { Field, SettingsSection } from "./SettingsSection";

export function GeneralSettingsForm({value,setValue,save,canEdit}:{value:GeneralSettings;setValue:(v:GeneralSettings)=>void;save:<K extends keyof GeneralSettings>(s:K)=>Promise<void>;canEdit:boolean}){
 const set=<K extends keyof GeneralSettings>(section:K,key:keyof GeneralSettings[K],next:unknown)=>setValue({...value,[section]:{...value[section],[key]:next}});
 return <>
  <SettingsSection id="settings-workspace" className="settings-card-wide settings-card-workspace" title="工作区" description="工作区身份、语言和时间显示规则。" canEdit={canEdit} onSave={()=>save("workspace")}>
   <Field label="名称"><input disabled={!canEdit} value={value.workspace.name} onChange={e=>set("workspace","name",e.target.value)}/></Field>
   <Field label="Slug" hint="用于公开 API 路径，修改前请确认外部链接。"><input disabled={!canEdit} value={value.workspace.slug} onChange={e=>set("workspace","slug",e.target.value)}/></Field>
   <Field label="默认语言"><select disabled={!canEdit} value={value.workspace.default_locale} onChange={e=>set("workspace","default_locale",e.target.value)}>{value.workspace.supported_locales.map(v=><option key={v}>{v}</option>)}</select></Field>
   <Field label="支持语言" hint="使用英文逗号分隔。"><input disabled={!canEdit} value={value.workspace.supported_locales.join(", ")} onChange={e=>set("workspace","supported_locales",e.target.value.split(",").map(v=>v.trim()).filter(Boolean))}/></Field>
   <Field label="时区"><input disabled={!canEdit} value={value.workspace.timezone} onChange={e=>set("workspace","timezone",e.target.value)}/></Field>
  </SettingsSection>
  <SettingsSection className="settings-card-wide settings-card-site" title="公开站点" description="控制站点介绍、公开地址和 RSS。" canEdit={canEdit} onSave={()=>save("site")}>
   <Field label="站点名称"><input disabled={!canEdit} value={value.site.name} onChange={e=>set("site","name",e.target.value)}/></Field>
   <Field label="公开地址"><input disabled={!canEdit} type="url" value={value.site.public_url} onChange={e=>set("site","public_url",e.target.value)}/></Field>
   <Field label="站点描述"><textarea disabled={!canEdit} value={value.site.description} onChange={e=>set("site","description",e.target.value)}/></Field>
   <Field label="关于页面"><textarea disabled={!canEdit} value={value.site.about} onChange={e=>set("site","about",e.target.value)}/></Field>
   <Field label="页脚"><input disabled={!canEdit} value={value.site.footer} onChange={e=>set("site","footer",e.target.value)}/></Field>
   <Field label="站点主题"><select disabled={!canEdit} value={value.site.theme||"paper"} onChange={e=>set("site","theme",e.target.value)}><option value="paper">纸张编辑风</option><option value="minimal">现代极简</option><option value="dark">深色阅读</option></select></Field>
   <Field label="强调色"><input disabled={!canEdit} type="color" value={value.site.accent_color||"#d82f76"} onChange={e=>set("site","accent_color",e.target.value)}/></Field>
   <Field label="分享海报底部文案"><input disabled={!canEdit} value={value.site.share_footer||""} onChange={e=>set("site","share_footer",e.target.value)} placeholder="扫码阅读完整内容"/></Field>
   <label className="settings-check"><input disabled={!canEdit} type="checkbox" checked={value.site.rss_enabled} onChange={e=>set("site","rss_enabled",e.target.checked)}/>启用 RSS</label>
  </SettingsSection>
  <SettingsSection id="settings-auth" className="settings-card-wide settings-card-auth" title="登录与发布" description="登录方式、会话期限和默认发布行为。会话期限应用于后续新会话。" canEdit={canEdit} onSave={async()=>{await save("auth");await save("publication")}}>
   <div className="settings-auth-methods" aria-label="允许的登录方式">
    <label className="settings-check"><input disabled={!canEdit} type="checkbox" checked={value.auth.password_login_enabled} onChange={e=>set("auth","password_login_enabled",e.target.checked)}/><span><strong>邮箱密码</strong><small>使用邮箱和密码登录</small></span></label>
    <label className="settings-check"><input disabled={!canEdit} type="checkbox" checked={value.auth.passkey_enabled} onChange={e=>set("auth","passkey_enabled",e.target.checked)}/><span><strong>Passkey</strong><small>使用设备生物识别或安全密钥</small></span></label>
    <label className="settings-check"><input disabled={!canEdit} type="checkbox" checked={value.auth.totp_login_enabled} onChange={e=>set("auth","totp_login_enabled",e.target.checked)}/><span><strong>TOTP 动态码</strong><small>使用邮箱和验证器动态码直接登录</small></span></label>
   </div>
   <div className="settings-auth-options">
    <Field label="会话有效期（小时）"><input disabled={!canEdit} type="number" min={1} max={8760} value={value.auth.session_ttl_hours} onChange={e=>set("auth","session_ttl_hours",Number(e.target.value))}/></Field>
    <Field label="默认发布渠道"><select disabled={!canEdit} value={value.publication.default_channel} onChange={e=>set("publication","default_channel",e.target.value)}><option value="website">网站</option><option value="rss">RSS</option></select></Field>
   </div>
  </SettingsSection>
 </>
}
