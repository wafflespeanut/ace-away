<template>
  <v-dialog :hide-overlay="true" v-model="showDialog" max-width="400" persistent>
    <v-card>
      <v-list-item>
        <v-card-title class="headline">{{ headline }} Room</v-card-title>
        <v-row align="center" justify="end">
          <v-switch v-model="isJoin" :class="{
            'mx-3': $vuetify.breakpoint.xs,
            'mx-5': $vuetify.breakpoint.smAndUp,
          }" label="Existing?"></v-switch>
        </v-row>
      </v-list-item>
      <v-card-text class="pl-8">{{ dialogText }}</v-card-text>
      <v-row class="mb-6" no-gutters>
        <v-col></v-col>
        <v-col :cols="8">
          <v-form v-model="formValid">
            <v-text-field v-model="player"
                          :rules="nameRules"
                          label="Your name" required></v-text-field>
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
  /**
   * Headline for the modal.
   */
  private get headline(): string {
    return this.isJoin ? 'Join' : 'Create';
  }

  /** Subtitle for the modal. */
  private get dialogText(): string {
    return this.isJoin ? 'Join an existing room.' : 'Create a new room for playing.';
  }

  /** Whether player input for this dialog is valid. */
  private formValid: boolean = false;

  /** Whether the player is trying to join a room rather than creating one. */
  private isJoin: boolean = false;

  /** Validation rules for player name. */
  private nameRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Name is required',
    (v) => (v && v.length >= 3 && v.length <= 10) || 'Name must be 3-10 characters',
  ];

  /** Room name. */
  private roomName: string = '';

  /** Player ID. */
  private player: string = '';

  /** Validation rules for room name. */
  private get roomRules(): Array<(v: any) => boolean | string> {
    if (!this.isJoin) {
      return [];
    }

    return [
      (v) => !!v || 'Room name required',
    ];
  }

  /** Number of players allowed in this room (for creation). */
  private numPlayers: number | null = null;

  /** Validation rules for player count. */
  private playersRules: Array<(v: any) => boolean | string> = [
    (v) => !!v || 'Player count is required',
  ];

  /** Loading of dialog button. */
  private loading: boolean = false;

  private conn = new ConnectionProvider();

  /**
   * Callback initiated once the player has created a room.
   */
  private roomCreated() {
    this.loading = true;
    // Emit an event to the parent with the room name.
    this.$emit('player-set', this.player.toLocaleLowerCase());

    this.conn.onError((msg, e) => {
      if (e === GameEvent.playerExists || e === GameEvent.roomMissing || e === GameEvent.roomExists) {
        this.loading = false;
      }
    });

    if (this.isJoin) {
      this.conn.joinRoom(this.player, this.roomName);
    } else {
      this.conn.createRoom(this.player, this.roomName, this.numPlayers!);
    }
  }
}
</script>
