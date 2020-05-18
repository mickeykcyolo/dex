import Joi from '@hapi/joi';

//@ts-ignore
const CustomJoi = Joi.extend(require('joi-phone-number'))



export const UserDataScheme = Joi.object({
    id: Joi.string().required(),
    kind: Joi.string().required(),
    name: Joi.string().required(),
    system: Joi.bool().required(),
    enabled: Joi.bool().required(),
    phone_number: Joi.string().allow(''),
    ctime: Joi.string().required(),
    mtime: Joi.string().required(),
    supervisor: Joi.any(),
    auth_level: Joi.number(),
    personal_desktop: Joi.string().allow(''),
    enrolled: Joi.bool().required(),
    totp_enabled: Joi.bool().required(),
    totp_enrolled: Joi.bool().required(),
})
export const PasswordScheme = Joi.string().min(2).max(64).required()
export const UsernameScheme = Joi.string().min(3).max(64).required()
export const PhoneNumberScheme = CustomJoi.string().phoneNumber().required()