import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse, GameEvent, DealResponse } from './model';

export default interface GameEventHub {

    createRoom(req: ClientMessage<RoomCreationRequest>): void;

    joinRoom(req: ClientMessage<{}>): void;

    onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean): void;

    onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean): void;

    onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean): void;
}
