import { Principal, request } from "./api";

type Ceremony = {
  ceremony_id: string;
  options: { publicKey: Record<string, unknown> };
};
type CredentialJSON = Credential & { toJSON?: () => unknown };
function creationOptions(value: Ceremony) {
  const api = PublicKeyCredential as unknown as {
    parseCreationOptionsFromJSON?: (
      v: unknown,
    ) => PublicKeyCredentialCreationOptions;
  };
  return (
    api.parseCreationOptionsFromJSON?.(value.options.publicKey) ??
    (value.options.publicKey as unknown as PublicKeyCredentialCreationOptions)
  );
}
function requestOptions(value: Ceremony) {
  const api = PublicKeyCredential as unknown as {
    parseRequestOptionsFromJSON?: (
      v: unknown,
    ) => PublicKeyCredentialRequestOptions;
  };
  return (
    api.parseRequestOptionsFromJSON?.(value.options.publicKey) ??
    (value.options.publicKey as unknown as PublicKeyCredentialRequestOptions)
  );
}
function json(value: CredentialJSON) {
  return value.toJSON?.() ?? value;
}
export async function loginWithPasskey(email: string) {
  const begin = await request<Ceremony>("/v1/auth/passkeys/login/options", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
  const credential = await navigator.credentials.get({
    publicKey: requestOptions(begin),
  });
  if (!credential) throw new Error("未选择 Passkey");
  return request<Principal>(
    `/v1/auth/passkeys/login/verify?ceremony_id=${begin.ceremony_id}`,
    { method: "POST", body: JSON.stringify(json(credential)) },
  );
}
export async function registerPasskey() {
  const begin = await request<Ceremony>("/v1/auth/passkeys/register/options", {
    method: "POST",
    body: "{}",
  });
  const credential = await navigator.credentials.create({
    publicKey: creationOptions(begin),
  });
  if (!credential) throw new Error("未创建 Passkey");
  await request<void>(
    `/v1/auth/passkeys/register/verify?ceremony_id=${begin.ceremony_id}`,
    { method: "POST", body: JSON.stringify(json(credential)) },
  );
}
