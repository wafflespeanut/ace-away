<template>
  <v-dialog v-model="showDialog" max-width="400" persistent>
    <v-card>
      <v-list-item>
        <v-card-title class="headline">{{ headline }} Room</v-card-title>
        <v-row align="center" justify="end">
          <v-switch v-model="isJoin" class="mx-5" label="Existing?"></v-switch>
        </v-row>
      </v-list-item>
      <v-card-text class="pl-8">{{ dialogText }}</v-card-text>
      <v-row class="mb-6" no-gutters>
        <v-col></v-col>
        <v-col :cols="8">
          <v-form v-model="formValid">
            <v-text-field v-model="player" :rules="nameRules" label="Your name" required></v-text-field>
            <v-text-field v-model="roomName" :rules="roomRules" label="Room name" required></v-text-field>
            <v-select v-if="!isJoin" v-model="numPlayers" :rules="playersRules" label="No. of players" :items="players" required></v-select>
          </v-form>
        </v-col>
        <v-col></v-col>
        <v-col></v-col>
      </v-row>
      <v-card-actions>
        <v-btn color="red"
               :loading="loading"
               :disabled="!formValid || loading"
               class="mx-auto my-5 px-5" depressed
               @click="roomCreated">{{ headline }}</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import ConnectionProvider from '../persistence/connection';
import { RoomCreationRequest, GameEvent } from '../persistence/model';

const JoinRoomProps = Vue.extend({
  props: {
    showDialog: Boolean,
    players: Array,
  },
});

@Component({
  components: {},
})
export default class JoinRoom extends JoinRoomProps {

  private get headline(): string {
    return this.isJoin ? 'Join' : 'Create';
  }

  private get dialogText(): string {
    return this.isJoin ? 'Join an existing room.' : 'Create a new room for playing.';
  }

  private formValid: boolean = false;

  private isJoin: boolean = false;

  private player: string = '';

  private nameRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Name is required',
    (v) => (v && v.length >= 3 && v.length <= 10) || 'Name must be 3-10 characters',
  ];

  private roomName: string = '';

  private get roomRules(): Array<(v: any) => boolean | string> {
    if (!this.isJoin) {
      return [];
    }

    return [
      (v) => !!v || 'Room name required',
    ];
  }

  private numPlayers: number | null = null;

  private playersRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Player count is required',
  ];

  private loading: boolean = false;

  private conn = new ConnectionProvider();

  /**
   * Callback initiated once the player has created a room.
   */
  private roomCreated() {
    this.loading = true;

    this.conn.onPlayerJoin((resp) => {
      this.$emit('joined', this.player, resp);
    });

    this.conn.onError((msg, e) => {
      if (e === GameEvent.playerExists || e === GameEvent.roomMissing || e === GameEvent.roomExists) {
        this.loading = false;
      }
    });

    if (this.isJoin) {
      this.conn.joinRoom({
        player: this.player,
        room: this.roomName,
        event: GameEvent.playerJoin,
        data: {},
      });
    } else {
      this.conn.createRoom({
        player: this.player,
        room: this.roomName,
        event: GameEvent.createRoom,
        data: {
          players: this.numPlayers!,
        },
      });
    }
  }
}
</script>
