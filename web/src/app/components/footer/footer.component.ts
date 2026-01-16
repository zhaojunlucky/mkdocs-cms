import {Component, OnInit} from '@angular/core';

import {SiteServiceService} from '../../services/site.service.service';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import packageInfo from '../../../../package.json';

@Component({
  selector: 'app-footer',
  standalone: true,
  imports: [
    MatToolbarModule,
    MatButtonModule,
    MatIconModule
  ],
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {
  version = 'Unknown';
  frontendVersion = packageInfo.version;

  constructor(private siteService: SiteServiceService) {
  }

  ngOnInit(): void {
    this.siteService.getVersion().subscribe(version => {
      this.version = version.version
    })
  }
}
