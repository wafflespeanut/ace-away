import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './model';

export default interface GameEventHub {

    createRoom(req: ClientMessage<RoomCreationRequest>): void;

    onPlayerJoin(callback: (resp: ServerMessage<RoomResponse>) => void): void;
}
