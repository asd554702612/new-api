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

const HomeCTA = (props) => {
  return (
    <section className='mx-auto mt-10 max-w-7xl'>
      <div className='relative overflow-hidden rounded-[2.5rem] bg-[linear-gradient(135deg,#020617,#0f172a_46%,#111827)] px-6 py-10 text-white shadow-[0_30px_100px_rgba(2,6,23,0.36)] sm:px-10 sm:py-14'>
        <div className='pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(59,130,246,0.28),transparent_30%),radial-gradient(circle_at_bottom_right,rgba(34,211,238,0.22),transparent_28%)]' />
        <div className='relative z-10 max-w-3xl'>
          <span className='inline-flex rounded-full border border-white/15 bg-white/10 px-4 py-2 text-sm font-semibold text-cyan-100'>
            {props.data.badge}
          </span>
          <h2
            className='mt-6 text-4xl font-black tracking-tight sm:text-5xl'
            style={{ color: '#fff' }}
          >
            {props.data.title}
          </h2>
          <p className='mt-5 max-w-2xl text-base leading-8 text-slate-300 sm:text-lg'>
            {props.data.subtitle}
          </p>

          <div className='mt-8 flex flex-col gap-4 sm:flex-row'>
            <Link
              to={props.consolePath}
              className='inline-flex items-center justify-center rounded-full bg-white px-8 py-4 text-base font-black text-slate-950 transition-transform hover:-translate-y-0.5 hover:bg-slate-100'
            >
              {props.primaryLabel}
            </Link>
            <a
              href={props.docUrl}
              target='_blank'
              rel='noopener noreferrer'
              className='inline-flex items-center justify-center rounded-full border border-white/20 bg-white/5 px-8 py-4 text-base font-bold text-white transition-transform hover:-translate-y-0.5 hover:bg-white/10'
            >
              {props.secondaryLabel}
            </a>
          </div>

          <div className='mt-8 flex flex-wrap gap-3 text-sm text-slate-300'>
            {props.data.trustItems.map((item) => (
              <span
                key={item}
                className='inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-4 py-2'
              >
                <span className='text-cyan-300'>●</span>
                {item}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default HomeCTA;
