export interface IUser {
    id: string,
    kind: string,
    name: string,
    system: boolean,
    enabled: boolean,
    phone_number: string,
    ctime: string,
    mtime: string,
    supervisor: string,
    auth_level: number,
    personal_desktop: string,
    enrolled: boolean,
    totp_enabled: boolean,
    totp_enrolled: boolean,
}