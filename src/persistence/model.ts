
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

interface TurnRequest {
  card: Card;
}

interface RoomResponse {
  players: string[];
  max: number;
  escaped: string[];
  turnIdx: number;
}

interface DealResponse {
  hand: Card[];
  table: PlayerCard[];
  isDealer: boolean;
  ourTurn: boolean;
  turnPlayer: string;
}

enum GameEvent {
  createRoom = 'RoomCreate',
  playerJoin = 'PlayerJoin',
  playerExists = 'PlayerExists',
  roomMissing = 'RoomMissing',
  roomExists = 'RoomExists',
  playerTurn = 'PlayerTurn',
  playerWins = 'PlayerWin',
  gameOver = 'GameOver',
}

interface PlayerCard {
  id: string;
  card: Card;
}

interface Card {
  label: Label;
  suite: Suite;
}

enum Suite {
  Diamond = 'd',
  Clover  = 'c',
  Heart   = 'h',
  Spade   = 's',
}

const suitePrettyMap: any = {
  d: '♦',
  c: '♣',
  h: '♥',
  s: '♠',
};

enum Label {
  Two   = '2',
  Three = '3',
  Four  = '4',
  Five  = '5',
  Six   = '6',
  Seven = '7',
  Eight = '8',
  Nine  = '9',
  Ten   = '10',
  Jack  = 'J',
  Queen = 'Q',
  King  = 'K',
  Ace   = 'A',
}

const suiteRanks: any = {
  d: 1,
  c: 10,
  h: 100,
  s: 1000,
};

const labelRanks: any = {
  2:  2,
  3:  3,
  4:  4,
  5:  5,
  6:  6,
  7:  7,
  8:  8,
  9:  9,
  10: 10,
  J:  11,
  Q:  12,
  K:  13,
  A:  14,
};

export {
  Card, Label, Suite, labelRanks, suiteRanks, suitePrettyMap,
  DealResponse, RoomResponse, GameEvent, RoomCreationRequest,
  ClientMessage, ServerMessage, PlayerCard, TurnRequest,
};
