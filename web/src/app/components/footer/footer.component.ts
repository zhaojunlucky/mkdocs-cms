import {Component, OnInit} from '@angular/core';

import {SiteServiceService} from '../../services/site.service.service';
import packageInfo from '../../../../package.json';

@Component({
  selector: 'app-footer',
  standalone: true,
  imports: [],
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {
  currentYear = new Date().getFullYear();
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
