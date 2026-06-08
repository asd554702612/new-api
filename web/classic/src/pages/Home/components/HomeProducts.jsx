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
import { Cpu, Globe, Sparkles } from 'lucide-react';

const iconMap = {
  cpu: Cpu,
  sparkles: Sparkles,
};

const HomeProducts = (props) => {
  return (
    <section className='overflow-hidden bg-gradient-to-b from-white to-zinc-50 py-32 dark:from-[#040816] dark:to-[#071022]'>
      <div className='relative mx-auto mb-48 max-w-7xl px-4 sm:px-6 lg:px-8'>
        <div className='relative z-10 mb-16 text-center'>
          <span className='mb-6 inline-block text-sm font-semibold uppercase tracking-widest text-zinc-500 dark:text-slate-400'>
            {props.data.kicker}
          </span>
          <p className='mx-auto mb-8 max-w-3xl text-xl font-medium tracking-wide text-zinc-600 dark:text-slate-300'>
            {props.data.subtitle}
          </p>
          <h2 className='flex flex-wrap items-center justify-center gap-6 font-serif text-5xl leading-tight tracking-tight text-zinc-900 dark:text-white md:text-7xl lg:text-8xl'>
            {props.data.title}
            <span className='relative inline-block'>
              <div className='group relative mx-4 h-56 w-56 md:h-72 md:w-72 lg:h-96 lg:w-96'>
                <div className='absolute inset-0 animate-[spin_40s_linear_infinite] rounded-full border-2 border-dashed border-zinc-200 opacity-60 dark:border-slate-700' />
                <div className='absolute inset-[15%] animate-[spin_25s_linear_infinite_reverse] rounded-full border border-zinc-100 dark:border-slate-800'>
                  <div className='absolute left-1/2 top-0 h-2 w-2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-blue-500 shadow-[0_0_10px_#3b82f6]' />
                  <div className='absolute bottom-0 left-1/2 h-2 w-2 -translate-x-1/2 translate-y-1/2 rounded-full bg-zinc-300 dark:bg-slate-600' />
                </div>
                <div className='absolute inset-[25%] flex items-center justify-center overflow-hidden rounded-full border border-zinc-200 bg-white shadow-[0_0_50px_rgba(0,0,0,0.03)] transition-transform duration-700 group-hover:scale-105 dark:border-slate-700 dark:bg-slate-900'>
                  <div className='absolute inset-0 opacity-[0.05] [background-image:linear-gradient(to_right,#000_1px,transparent_1px),linear-gradient(to_bottom,#000_1px,transparent_1px)] [background-size:25px_25px]' />
                  <svg
                    viewBox='0 0 800 400'
                    className='absolute inset-0 h-full w-full scale-150 object-contain opacity-[0.08] animate-[pulse_4s_ease-in-out_infinite]'
                    aria-hidden='true'
                  >
                    <path
                      fill='currentColor'
                      d='M150,150 L160,140 L180,145 L190,160 L180,175 L160,170 Z M300,100 L320,90 L350,110 L340,130 L310,120 Z M500,200 L530,190 L560,210 L540,240 L510,230 Z M650,80 L680,70 L700,90 L670,110 Z M200,300 L230,290 L250,310 L220,330 Z'
                    />
                  </svg>
                  <div className='pointer-events-none absolute inset-0 flex items-center justify-center'>
                    <div className='absolute h-full w-full animate-[ping_3s_linear_infinite] rounded-full border border-blue-400/20' />
                    <div className='absolute h-3/4 w-3/4 animate-[ping_4s_linear_infinite_1s] rounded-full border border-blue-400/10' />
                  </div>
                  <div className='relative z-10 rounded-full border border-zinc-100 bg-white p-5 shadow-sm dark:border-slate-700 dark:bg-slate-950'>
                    <Globe
                      size={48}
                      className='text-zinc-800 dark:text-white'
                    />
                  </div>
                </div>
              </div>
            </span>
          </h2>
        </div>

        <div className='mx-auto mt-24 grid max-w-5xl grid-cols-1 gap-8 md:grid-cols-2'>
          {props.data.cards.map((card) => {
            const Icon = iconMap[card.icon];
            return (
              <article
                key={card.title}
                className='rounded-3xl border-2 border-zinc-200/60 bg-white p-8 shadow-lg transition-all duration-500 hover:border-zinc-300 hover:shadow-2xl dark:border-slate-800/80 dark:bg-slate-900/78 dark:hover:border-slate-600'
              >
                <div className='flex gap-5'>
                  <div className='flex h-14 w-14 shrink-0 items-center justify-center rounded-xl border border-blue-200/50 bg-gradient-to-br from-blue-100 to-blue-50 text-blue-600 dark:border-slate-700 dark:from-slate-800 dark:to-slate-900 dark:text-cyan-200'>
                    <Icon size={28} />
                  </div>
                  <div>
                    <div className='mb-3 inline-block rounded-lg bg-black px-3 py-1 text-xs font-bold text-white dark:bg-cyan-300 dark:text-slate-950'>
                      {card.title}
                    </div>
                    <p className='text-sm leading-relaxed text-zinc-700 dark:text-slate-300'>
                      {card.description}
                    </p>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      </div>

      <div className='relative mx-auto max-w-7xl border-t-2 border-zinc-200/60 bg-gradient-to-b from-zinc-50 to-white px-4 py-24 sm:px-6 lg:px-8 dark:border-slate-800 dark:from-[#071022] dark:to-[#040816]'>
        <div className='relative z-10 text-center'>
          <div className='mb-8 inline-flex h-20 w-20 items-center justify-center rounded-2xl bg-black text-white shadow-2xl transition-all duration-500 hover:rotate-6 hover:scale-110 dark:bg-cyan-300 dark:text-slate-950'>
            <Sparkles size={40} />
          </div>
          <h3 className='mb-4 font-serif text-4xl font-bold leading-tight text-zinc-900 dark:text-white md:text-5xl'>
            {props.data.sideTitle}
          </h3>
          <p className='mb-12 text-xl font-medium text-zinc-600 dark:text-slate-300'>
            {props.data.sideSubtitle}
          </p>

          <div className='mx-auto grid w-full max-w-5xl grid-cols-2 justify-items-center gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5'>
            {props.data.models.map((item, index) => (
              <div
                key={item}
                className={`flex min-h-16 w-full max-w-[13rem] items-center justify-center rounded-2xl px-4 py-3 text-center text-sm font-semibold leading-snug shadow-sm transition-colors sm:text-base ${
                  index === 0
                    ? 'bg-black text-white shadow-lg dark:bg-cyan-300 dark:text-slate-950'
                    : 'border-2 border-zinc-200 bg-white text-black dark:border-slate-700 dark:bg-slate-950 dark:text-white'
                }`}
              >
                {item}
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default HomeProducts;
