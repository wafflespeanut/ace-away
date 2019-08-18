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
      <v-container class="fill-height" fluid>
        <!--  -->
      </v-container>
      <JoinRoom @joined="playerJoined" :players="allowedPlayers" :showDialog="!roomJoined" />
    </v-content>
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

  private playerJoined(player: string, resp: ServerMessage<RoomResponse>) {
    if (resp.player === player) {
      this.roomJoined = true;
    }
  }
}
export { ALLOWED_PLAYERS };
</script>
