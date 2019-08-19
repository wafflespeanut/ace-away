<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawerOpen" app>
      <!--  -->
    </v-navigation-drawer>

    <v-app-bar app>
      <v-app-bar-nav-icon @click.stop="drawerOpen = !drawerOpen"></v-app-bar-nav-icon>
      <v-toolbar-title>Ace Away!</v-toolbar-title>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <v-alert v-if="alertMsg" border="top" colored-border type="info" elevation="2">{{ alertMsg }}</v-alert>
      </v-container>
      <JoinRoom @error="showError" @joined="playerJoined" :players="allowedPlayers" :showDialog="!roomJoined" />
    </v-content>
    <v-snackbar :value="true" :timeout="0" v-for="(text, i) in notifications" :key="i">
      {{ text }}
      <v-btn color="pink" text @click="notifications.splice(i, 1)">Close</v-btn>
    </v-snackbar>
    <v-footer app>
      <span class="white--text">&copy; 2019 @wafflespeanut</span>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';
import JoinRoom from './dialog/JoinRoom.vue';
import { ClientMessage, RoomCreationRequest, ServerMessage, RoomResponse } from './persistence/model';
import ConnectionProvider from './persistence/connection';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];

@Component({
  components: {
    JoinRoom,
  },
})
export default class App extends Vue {

  private drawerOpen: boolean = false;

  private roomJoined: boolean = false;

  private allowedPlayers: number[] = ALLOWED_PLAYERS;

  private notifications: string[] = [];

  private alertMsg: string | null = null;

  private playerJoined(player: string, resp: ServerMessage<RoomResponse>) {
    if (resp.player === player) {
      this.roomJoined = true;
    }

    const diff = resp.response.max - resp.response.players.length;
    if (diff > 0) {
      this.alertMsg = `Waiting for ${diff} more player(s).`;
    } else {
      this.alertMsg = null;
    }
  }

  private showError(msg: string) {
    this.notifications.push(msg);
  }
}
export { ALLOWED_PLAYERS };
</script>
