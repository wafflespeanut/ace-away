<template>
  <v-dialog v-model="showDialog" max-width="400" persistent>
    <v-card>
      <v-card-title class="headline">Create Room</v-card-title>
      <v-card-text>Create a new room for playing.</v-card-text>
      <v-row class="mb-6" no-gutters>
        <v-col></v-col>
        <v-col :cols="8">
          <v-form v-model="formValid">
            <v-text-field v-model="player" :rules="nameRules" label="Your name" required></v-text-field>
            <v-text-field v-model="roomName" label="Room name (optional)"></v-text-field>
            <v-select v-model="numPlayers" :rules="playersRules" label="No. of players" :items="players" required></v-select>
          </v-form>
        </v-col>
        <v-col></v-col>
        <v-col></v-col>
      </v-row>
      <v-card-actions>
        <v-btn color="red"
               :loading="loading"
               :disabled="!formValid || loading"
               class="mx-auto px-5" depressed
               @click="roomCreated">Create</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import ConnectionProvider from '../persistence/connection';
import { RoomCreationRequest, GameEvent } from '../persistence/model';

const CreateRoomProps = Vue.extend({
  props: {
    showDialog: Boolean,
    players: Array,
  },
});

@Component({
  components: {},
})
export default class CreateRoom extends CreateRoomProps {

  private formValid: boolean = false;

  private player: string = '';

  private roomName: string = '';

  private numPlayers: number | null = null;

  private nameRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Name is required',
    (v) => (v && v.length >= 3 && v.length <= 10) || 'Name must be 3-10 characters',
  ];

  private playersRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Player count is required',
  ];

  private loading: boolean = false;

  private conn = new ConnectionProvider();

  /**
   * Callback initiated once the user has created a room.
   */
  private roomCreated() {
    this.loading = true;

    this.conn.onPlayerJoin((resp) => {
      if (resp.player === this.player && this.loading) {
        console.log(`Joined room ${resp.room}`);
        this.$props.showDialog = false;
      }
    });

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
</script>
