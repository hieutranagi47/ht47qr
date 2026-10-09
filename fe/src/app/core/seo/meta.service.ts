import { APP_BASE_HREF } from '@angular/common';
import { Injectable, inject, DOCUMENT } from '@angular/core';
import { Meta, Title } from '@angular/platform-browser';
import { Router } from '@angular/router';
import { environment } from '@env/environment';

const DEFAULT = {
  title: "Hieu Tran's Profile",
  description:
    'Experience with Head of Development, Front-end technical leader, Front-end architecture, Team management, and Project management. Good at Communication Skills',
  imgUrl: `${environment.host}/assets/imgs/avatar.jpg`,
  keywords:
    'Front-end developer, Angular, ReactJS/Next, Vue/Nuxt, SolidJS/SolidStart, AngularJS, HTML5, CSS3, SCSS, LESS, SASS, Boostrap 5, AntDesign, Python - FastApi, Falcon, Flask, PHP - Zend 2 framework, Go (Golang) - Gin, Mongo DB, MySQL, MS SQL, Rest API, (g)PRC, GraphQL, MicroService, Micro Front-end, Design System, Storybook, Atomic Design',
  imageDimension: { default: 256, og: 512 }, // og for Facebook rule is min size of image 200 x 200
};

@Injectable({
  providedIn: 'root',
})
export class SeoMetaService {
  dom = inject(DOCUMENT);
  private baseHref = inject(APP_BASE_HREF);
  private title = inject(Title);
  private meta = inject(Meta);
  private router = inject(Router);

  private getImgUrlByDimension(imgUrl: string, dimension: number): string {
    const extensionpartPos = imgUrl.lastIndexOf('.');
    const socialImgUrl = `${imgUrl.substring(
      0,
      extensionpartPos
    )}_${dimension}${imgUrl.substring(extensionpartPos)}`;

    return socialImgUrl;
  }

  private getSocialImage(
    imgUrl: string,
    dimensions: number[],
    dimensionVal: number
  ): string {
    let dimension: number = 0;
    for (const i of dimensions) {
      if (i === dimensionVal) {
        dimension = i;
        break;
      }
    }

    if (!dimension) {
      return imgUrl;
    }

    return this.getImgUrlByDimension(imgUrl, dimension);
  }

  /**
   * Update meta tag for page.
   * @param title Meta tag title
   * @param desc Meta tag description
   * @param imgUrl Meta tag image
   * @param options Options for SEO.
   */
  updateMeta(
    title: string = DEFAULT.title,
    desc: string = DEFAULT.description,
    keywords: string = DEFAULT.keywords,
    imgUrl: string = DEFAULT.imgUrl,
    options: any = {}
  ): void {
    const titleName =
      title === DEFAULT.title ? title : title + " | Hieu Tran's Profile";

    let socialImgUrl = imgUrl;
    let ogImgUrl = imgUrl;

    if (options.imgDimensions && options.imgDimensions.length > 0) {
      socialImgUrl = this.getSocialImage(
        imgUrl,
        options.imgDimensions,
        DEFAULT.imageDimension.default
      );
      ogImgUrl = this.getSocialImage(
        imgUrl,
        options.imgDimensions,
        DEFAULT.imageDimension.og
      );
    }

    // Canonical URL
    const baseUrl = environment.host + (this.baseHref === '/' ? '' : this.baseHref);
    const canonicalUrl = baseUrl + (this.router.url === '/' ? '' : this.router.url.substring(1));

    this.meta.updateTag({
      name: 'description',
      content: desc,
    });
    this.title.setTitle(titleName);
    this.meta.updateTag({ property: 'og:site_name', content: titleName });
    this.meta.updateTag({ property: 'og:title', content: titleName });
    this.meta.updateTag({ property: 'og:image:alt', content: titleName });
    this.meta.updateTag({
      property: 'og:description',
      content: desc,
    });

    // console.log('ogImgUrl--> ', ogImgUrl);
    this.meta.updateTag({
      property: 'og:image',
      content: ogImgUrl,
    });

    this.meta.updateTag({
      property: 'og:image:secure_url',
      content: ogImgUrl,
    });
    this.meta.updateTag({
      property: 'og:url',
      content: canonicalUrl,
    });

    this.meta.updateTag({ name: 'twitter:title', content: titleName });
    this.meta.updateTag({
      property: 'description',
      content: desc,
    });
    this.meta.updateTag({
      name: 'twitter:description',
      content: desc,
    });

    this.meta.updateTag({
      name: 'twitter:image',
      content: socialImgUrl,
    });
    this.meta.updateTag({
      name: 'twitter:image:src',
      content: socialImgUrl,
    });
    this.meta.updateTag({
      name: 'twitter:url',
      content: canonicalUrl,
    });
    this.meta.updateTag({
      name: 'twitter:card',
      content: 'summary_large_image',
    });
    this.meta.updateTag({
      name: 'keywords',
      content: keywords,
    });

    const canonical = this.dom.querySelector('link[rel=canonical]')!;
    canonical.setAttribute('href', canonicalUrl);
  }
}
