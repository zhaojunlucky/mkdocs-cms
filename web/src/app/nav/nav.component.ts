import { Component } from '@angular/core';

import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { RouterModule } from '@angular/router';
import { AuthService } from '../auth/auth.service';
import {MatDivider} from '@angular/material/divider';

@Component({
  selector: 'app-nav',
  standalone: true,
  imports: [
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    RouterModule,
    MatDivider
],
  templateUrl: './nav.component.html',
  styleUrls: ['./nav.component.scss']
})
export class NavComponent {
  constructor(public authService: AuthService) {}

  logout(): void {
    this.authService.logout();
  }

  showReqNotifyBtn(): boolean {
    return Notification.permission !== 'granted';
  }

  requestNotificationPermission(): void {
    if ('Notification' in window) {
      Notification.requestPermission().then(permission => {
        if (permission === 'granted') {
          new Notification('Notifications Enabled!', {
            body: 'You will now receive notifications for unsaved changes.',
            icon: '/favicon.ico'
          });
        }
      });
    }
  }
}
