import {Component, OnInit} from '@angular/core';
import { CommonModule } from '@angular/common';
import {SiteServiceService} from '../../services/site.service.service';

@Component({
  selector: 'app-footer',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {
  currentYear = new Date().getFullYear();
  version = 'Unknown';

  constructor(private siteService: SiteServiceService) {
  }

  ngOnInit(): void {
    this.siteService.getVersion().subscribe(version => {
      this.version = version.version
    })
  }
}
