import Config from '../config';

//TODO: move to DI pattern
export const build = (endpoint : string) => {
    return `${window.origin}/${Config.ApiVersion}/${endpoint}`
}