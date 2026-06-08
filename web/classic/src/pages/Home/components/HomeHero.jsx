/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';
import { Link } from 'react-router-dom';

const heroStreams = [
  { left: '8%', top: '-6%', animationDuration: '15s', animationDelay: '0s' },
  { left: '34%', top: '-12%', animationDuration: '20s', animationDelay: '4s' },
  { left: '62%', top: '-8%', animationDuration: '24s', animationDelay: '8s' },
];

const heroParticles = [
  { left: '8%', top: '18%', animationDuration: '20s', animationDelay: '0s' },
  { left: '16%', top: '62%', animationDuration: '22s', animationDelay: '2s' },
  { left: '28%', top: '34%', animationDuration: '24s', animationDelay: '5s' },
  { left: '44%', top: '70%', animationDuration: '26s', animationDelay: '3s' },
  { left: '58%', top: '20%', animationDuration: '21s', animationDelay: '7s' },
  { left: '72%', top: '52%', animationDuration: '25s', animationDelay: '1s' },
  { left: '84%', top: '26%', animationDuration: '23s', animationDelay: '6s' },
  { left: '90%', top: '76%', animationDuration: '27s', animationDelay: '4s' },
];

const HomeHero = (props) => {
  return (
    <section
      id='top'
      className='relative flex flex-col items-center justify-between overflow-hidden px-4 pb-[7vh] pt-[max(3rem,6vh)] sm:px-6 lg:px-8'
      style={{ minHeight: `calc(100vh - ${props.headerOffset || 120}px)` }}
      data-test='home-hero'
    >
      <div className='pointer-events-none absolute inset-0 overflow-hidden'>
        <div className='absolute inset-0 opacity-[0.03] [background-image:linear-gradient(to_right,#000_1px,transparent_1px),linear-gradient(to_bottom,#000_1px,transparent_1px)] [background-size:60px_60px] dark:opacity-[0.06]' />
        <div className='classic-home-landing-scan absolute inset-x-0 top-[-22%] h-[300px] bg-gradient-to-b from-transparent via-blue-500/[0.03] to-transparent dark:via-cyan-400/[0.06]' />
        {heroStreams.map((stream, index) => (
          <div
            key={index}
            className='classic-home-landing-stream absolute left-0 top-0 h-[420px] w-px origin-top rotate-45 bg-gradient-to-b from-transparent via-zinc-400/80 to-transparent dark:via-cyan-200/70'
            style={stream}
          />
        ))}
        <div className='classic-home-landing-glow absolute left-[10%] top-[-10%] h-[800px] w-[800px] rounded-full bg-blue-50/25 dark:bg-cyan-500/12' />
        <div className='classic-home-landing-glow classic-home-landing-glow-right absolute bottom-[-10%] right-[10%] h-[900px] w-[900px] rounded-full bg-zinc-100/40 dark:bg-blue-500/12' />
        {heroParticles.map((particle, index) => (
          <div
            key={index}
            className='classic-home-landing-particle absolute h-[1.5px] w-[1.5px] rounded-full bg-zinc-300 dark:bg-cyan-100/70'
            style={particle}
          />
        ))}
      </div>

      <div className='relative z-10 mx-auto flex w-full max-w-7xl flex-1 flex-col items-center justify-center gap-10 py-6 lg:min-h-0 lg:flex-row lg:items-center lg:gap-16 lg:py-10'>
        <div className='flex w-full flex-1 flex-col justify-center text-center lg:text-left'>
          <div className='mb-8 flex flex-col items-center justify-center gap-5 md:flex-row lg:justify-start'>
            <div className='classic-home-landing-logo-float relative'>
              <div className='absolute inset-0 rounded-full bg-blue-400/10 blur-2xl dark:bg-cyan-400/18' />
              <div className='relative flex h-16 w-16 items-center justify-center rounded-full border border-zinc-200/80 bg-white shadow-sm dark:border-slate-700/80 dark:bg-slate-900'>
                <span className='text-lg font-black text-black dark:text-white'>
                  AI
                </span>
              </div>
            </div>
            <h1 className='flex min-h-[1.15em] items-center text-6xl font-serif tracking-[-0.08em] text-black sm:text-7xl md:text-8xl lg:text-[clamp(5.5rem,8vw,8rem)] dark:text-white'>
              <span className='inline-block bg-gradient-to-b from-black to-zinc-700 bg-clip-text text-transparent dark:from-white dark:to-slate-300'>
                {props.data.titlePrefix}
              </span>
            </h1>
          </div>

          <p className='mb-5 text-[12px] font-semibold uppercase tracking-[0.28em] text-zinc-400 dark:text-slate-400'>
            {props.data.badge}
          </p>

          <h2 className='mx-auto max-w-4xl text-4xl font-black leading-[1.04] tracking-[-0.055em] text-zinc-900 sm:text-5xl lg:mx-0 lg:max-w-3xl lg:text-[clamp(3.6rem,4.8vw,5.8rem)] dark:text-white'>
            {props.data.titleLead}
            <span className='text-blue-600 dark:text-cyan-300'>
              {props.data.titleHighlight}
            </span>
            {props.data.titleSuffix}
          </h2>

          <p className='mx-auto mt-7 max-w-2xl text-base font-light leading-relaxed tracking-wide text-zinc-500 sm:text-lg lg:mx-0 lg:text-xl dark:text-slate-300'>
            <span className='opacity-60'>—</span> {props.data.subtitle}{' '}
            <span className='opacity-60'>—</span>
          </p>

          <div className='mt-10 grid max-w-xl grid-cols-2 gap-4 md:gap-6 lg:mx-0'>
            {props.data.stats.map((item) => (
              <div
                key={item.label}
                className='group relative rounded-2xl border border-zinc-100/80 bg-white/60 px-5 py-5 transition-all duration-500 hover:-translate-y-1 hover:border-zinc-200 hover:shadow-[0_8px_30px_rgb(0,0,0,0.02)] md:px-6 md:py-6 dark:border-slate-800/80 dark:bg-slate-900/65'
              >
                <div className='origin-left text-2xl font-serif text-black transition-transform duration-500 group-hover:scale-105 md:text-3xl dark:text-white'>
                  {item.value}
                </div>
                <div className='mt-1 text-[10px] font-bold uppercase tracking-[0.2em] text-zinc-400 dark:text-slate-400'>
                  {item.label}
                </div>
              </div>
            ))}
          </div>

          <div className='mt-10 flex flex-col items-center gap-4 sm:flex-row lg:items-start'>
            <Link
              to={props.consolePath}
              className='group inline-flex items-center justify-center rounded-full bg-black px-8 py-4 text-base font-semibold text-white shadow-xl transition-all duration-300 hover:scale-105 hover:bg-zinc-800 dark:border dark:border-cyan-400/30 dark:bg-cyan-300 dark:text-slate-950 dark:hover:bg-cyan-200'
            >
              {props.primaryLabel}
            </Link>
            <a
              href={props.docUrl}
              target='_blank'
              rel='noopener noreferrer'
              className='inline-flex items-center justify-center rounded-full border-2 border-zinc-200 bg-white px-8 py-4 text-base font-semibold text-black transition-all duration-300 hover:scale-105 hover:border-zinc-400 dark:border-slate-700 dark:bg-slate-950 dark:text-white dark:hover:border-cyan-400/40'
            >
              {props.data.secondary}
            </a>
          </div>
        </div>

        <div
          className='relative flex w-full flex-1 items-center justify-center lg:min-h-[clamp(30rem,68vh,46rem)]'
          data-test='hero-visual-stack'
        >
          <div className='relative w-full max-w-[min(92vw,44rem)] lg:max-w-[min(48vw,44rem)]'>
            <div className='relative aspect-[4/3] w-full lg:aspect-square'>
              <div className='relative z-20 h-full w-full overflow-hidden rounded-[2rem] border border-zinc-200/60 bg-zinc-50 shadow-2xl dark:border-slate-800/80 dark:bg-slate-900'>
                <picture className='block h-full w-full'>
                  <source
                    srcSet='/home/home-hero-main.avif'
                    type='image/avif'
                  />
                  <source
                    srcSet='/home/home-hero-main.webp'
                    type='image/webp'
                  />
                  <img
                    src='/home/home-hero-main.png'
                    alt='Homepage hero visual'
                    width='1275'
                    height='1234'
                    className='block h-full w-full object-cover'
                    fetchPriority='high'
                  />
                </picture>
                <div className='pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.08),rgba(15,23,42,0.14))]' />
                <div className='pointer-events-none absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-slate-950/20 via-slate-900/10 to-transparent' />
              </div>

              <div className='classic-home-landing-card-float absolute -right-[6%] -top-[8%] z-30 hidden w-[46%] overflow-hidden rounded-2xl border border-zinc-200/50 bg-white shadow-xl md:block dark:border-slate-800/80 dark:bg-slate-900'>
                <div className='relative aspect-video'>
                  <picture className='block h-full w-full'>
                    <source
                      srcSet='/home/home-hero-side-top.avif'
                      type='image/avif'
                    />
                    <source
                      srcSet='/home/home-hero-side-top.webp'
                      type='image/webp'
                    />
                    <img
                      src='/home/home-hero-side-top.png'
                      alt='Hero supporting visual top'
                      width='1254'
                      height='1254'
                      className='block h-full w-full object-cover'
                    />
                  </picture>
                  <div className='pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.02),rgba(15,23,42,0.12))]' />
                </div>
              </div>

              <div className='classic-home-landing-card-float-alt absolute -bottom-[7%] -left-[5%] z-10 hidden w-[38%] overflow-hidden rounded-2xl border border-zinc-200/50 bg-white shadow-lg md:block dark:border-slate-800/80 dark:bg-slate-900'>
                <div className='relative aspect-square'>
                  <picture className='block h-full w-full'>
                    <source
                      srcSet='/home/home-hero-side-bottom.avif'
                      type='image/avif'
                    />
                    <source
                      srcSet='/home/home-hero-side-bottom.webp'
                      type='image/webp'
                    />
                    <img
                      src='/home/home-hero-side-bottom.png'
                      alt='Hero supporting visual bottom'
                      width='1254'
                      height='1254'
                      className='block h-full w-full object-cover'
                    />
                  </picture>
                  <div className='pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(15,23,42,0.04),rgba(15,23,42,0.12))]' />
                </div>
              </div>

              <div className='pointer-events-none absolute inset-0 rounded-full bg-blue-500/5 blur-[120px] dark:bg-cyan-400/10' />
            </div>
          </div>
        </div>
      </div>

      <div className='relative z-10 mt-8 hidden w-full border-t border-zinc-100/80 pt-8 md:block dark:border-slate-800/80'>
        <div className='mx-auto flex max-w-6xl items-center justify-center'>
          <div className='classic-home-landing-scroll-indicator flex flex-col items-center gap-3 text-zinc-300 dark:text-slate-600'>
            <div className='h-12 w-px bg-gradient-to-b from-zinc-200 to-transparent dark:from-slate-700' />
          </div>
        </div>
      </div>
    </section>
  );
};

export default HomeHero;
