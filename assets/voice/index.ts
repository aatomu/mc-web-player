type DiscordPayload = {
  cmd?: DiscordPayloadCmd;
  nonce: string;
  evt?: DiscordPayloadEvt;
  data?: any;
  args?: any;
};

type DiscordPayloadCmd = "DISPATCH" | "AUTHORIZE" | "AUTHENTICATE" | "GET_GUILD" | "GET_GUILDS" | "GET_CHANNEL" | "GET_CHANNELS" | "SUBSCRIBE" | "UNSUBSCRIBE" | "SET_USER_VOICE_SETTINGS" | "SELECT_VOICE_CHANNEL" | "GET_SELECTED_VOICE_CHANNEL" | "SELECT_TEXT_CHANNEL" | "GET_VOICE_SETTINGS" | "SET_VOICE_SETTINGS" | "SET_CERTIFIED_DEVICES" | "SET_ACTIVITY" | "SEND_ACTIVITY_JOIN_INVITE" | "CLOSE_ACTIVITY_REQUEST";
type DiscordPayloadEvt = "READY" | "ERROR" | "GUILD_STATUS" | "GUILD_CREATE" | "CHANNEL_CREATE" | "VOICE_CHANNEL_SELECT" | "VOICE_STATE_CREATE" | "VOICE_STATE_UPDATE" | "VOICE_STATE_DELETE" | "VOICE_SETTINGS_UPDATE" | "VOICE_CONNECTION_STATUS" | "SPEAKING_START" | "SPEAKING_STOP" | "MESSAGE_CREATE" | "MESSAGE_UPDATE" | "MESSAGE_DELETE" | "NOTIFICATION_CREATE" | "ACTIVITY_JOIN" | "ACTIVITY_SPECTATE" | "ACTIVITY_JOIN_REQUEST";

type User = {
  id: number;
  username: string;
  global_name: string;
  avatar: string;
  avatar_decoration_data: any;
  bot: string;
  flags: number;
  premium_type: number;
};

type VoiceState = {
  mute: boolean;
  deaf: boolean;
  self_mute: boolean;
  self_deaf: boolean;
  suppress: boolean;
};
// /**
//  * @typedef message
//  * @type {object}
//  * @property {User} author
//  * @property {string} author_color
//  * @property {boolean} bot
//  * @property {string} content
//  * @property {object} avatar_decoration_data
//  * @property {string} nick
//  * @property {string} timestamp
//  */
