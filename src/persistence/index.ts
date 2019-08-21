import {
    ClientMessage, RoomCreationRequest, ServerMessage,
    RoomResponse, GameEvent, DealResponse, TurnRequest,
} from './model';

export default interface GameEventHub {

    createRoom(req: ClientMessage<RoomCreationRequest>): void;

    joinRoom(req: ClientMessage<{}>): void;

    showCard(req: ClientMessage<TurnRequest>): void;

    onError(callback: (msg: string, event: GameEvent) => void, persist?: boolean): void;

    onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void, persist?: boolean): void;

    onPlayerTurn(callback: (resp: ServerMessage<DealResponse>) => void, persist?: boolean): void;

    onPlayerWin(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;

    onGameOver(callback: (resp: ServerMessage<{}>) => void, persist?: boolean): void;
}
