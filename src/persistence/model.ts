import { Card } from '@/deck';

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

interface DealResponse {
  hand: Card[];
  isDealer: boolean;
}

enum GameEvent {
  createRoom = 'RoomCreate',
  playerJoin = 'PlayerJoin',
  playerExists = 'PlayerExists',
  roomMissing = 'RoomMissing',
  roomExists = 'RoomExists',
  gameStart = 'GameBegins',
}

export { ClientMessage, DealResponse, RoomCreationRequest, RoomResponse, GameEvent, ServerMessage };
