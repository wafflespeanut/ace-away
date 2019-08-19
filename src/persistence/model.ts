interface ClientMessage<T> {
  player: string;
  room: string;
  event: GameEvent;
  data: T;
}

interface ServerMessage<T> {
  player: string;
  room: string;
  msg: string;
  event: GameEvent;
  response: T;
}

interface RoomCreationRequest {
  players: number;
}

interface RoomResponse {
  players: string[];
  max: number;
}

enum GameEvent {
  createRoom = 'RoomCreate',
  playerJoin = 'PlayerJoin',
  playerExists = 'PlayerExists',
  roomMissing = 'RoomMissing',
  roomExists = 'RoomExists',
}

export { ClientMessage, RoomCreationRequest, RoomResponse, GameEvent, ServerMessage };
