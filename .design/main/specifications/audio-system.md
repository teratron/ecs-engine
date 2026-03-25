# Audio System
**Version:** 0.1.0
**Status:** Draft
**Layer:** concept

## Overview
The audio system provides component-driven sound playback integrated with the ECS world. Each audio source is an asset; playback is controlled through components attached to entities. Spatial audio positions sound relative to a listener entity in 3D space. A backend abstraction allows the actual mixing and output to be handled by pluggable audio engines.

## Related Specifications
- [Asset System](asset-system.md)
- [App Framework](app-framework.md)

## 1. Motivation
Games require both global (music, UI sounds) and positional (footsteps, explosions) audio. Representing playback as ECS components means audio lifetime is tied to entity lifetime, enabling patterns like "despawn the bullet entity and its sound stops automatically." A backend interface avoids hard-coupling to any single audio library.

## 2. Constraints & Assumptions
- Audio mixing runs on a dedicated thread managed by the backend; the ECS side only issues commands.
- Asset loading is asynchronous — an `AudioPlayer` referencing an unloaded asset produces silence until the asset is ready.
- Spatial audio requires a `Transform` component on both the source entity and the listener entity.
- At most one `SpatialListener` entity may be active at a time.
- The system supports common sample rates (22 050, 44 100, 48 000 Hz) but does not resample at runtime — the backend is responsible for sample-rate conversion.

## 3. Core Invariants
1. An `AudioSink` is created automatically when an `AudioPlayer` begins playback; user code never constructs sinks directly.
2. Removing the `AudioPlayer` component from an entity stops and drops its associated sink.
3. `GlobalVolume` is applied multiplicatively to every sink's individual volume.
4. Spatial attenuation is computed every frame from the current `Transform` of source and listener.
5. A sink in the `DESPAWN` playback mode must trigger entity despawn exactly once upon completion.

## 4. Detailed Design

### 4.1 AudioSource Asset
An `AudioSource` holds decoded PCM sample data loaded from disk. Supported container formats are WAV, OGG/Vorbis, FLAC, and MP3. The asset system resolves format via file extension and delegates to the matching `AssetLoader`.

### 4.2 AudioPlayer Component
Attached to an entity to request playback of an `AudioSource` asset handle.

Fields (pseudo-code):
```
AudioPlayer {
    source:   Handle<AudioSource>
    settings: PlaybackSettings
}
```

### 4.3 PlaybackSettings
```
PlaybackSettings {
    mode:    PlaybackMode   // LOOP | ONCE | DESPAWN
    volume:  f32            // 0.0 .. 1.0, default 1.0
    speed:   f32            // playback rate multiplier, default 1.0
    spatial: bool           // if true, requires Transform
}
```

### 4.4 AudioSink / SpatialAudioSink
`AudioSink` is a control handle inserted by the audio system after playback begins. It exposes pause, resume, stop, and runtime volume/speed adjustment. When `spatial` is true, a `SpatialAudioSink` is inserted instead, adding distance-based attenuation and stereo panning derived from the relative position to the `SpatialListener`.

### 4.5 SpatialListener
A marker component placed on exactly one entity (typically the camera or player). The audio system queries its `Transform` each frame to compute relative positions for all spatial sources.

### 4.6 GlobalVolume Resource
A world resource holding a single `f32` master volume applied to all sinks before output. Defaults to 1.0.

### 4.7 AudioBackend Interface
```
AudioBackend {
    fn create_sink(settings, source_data) -> SinkHandle
    fn update_sink(handle, params)
    fn drop_sink(handle)
    fn set_master_volume(volume)
}
```
A default backend is provided. Alternative backends (e.g., web audio, console-specific) implement the same interface and are registered at app build time.

### 4.8 System Pipeline
1. **audio_added** — detects new `AudioPlayer` components, creates sinks via the backend.
2. **audio_control_sync** — propagates `AudioSink` mutations (pause/resume/volume) to the backend.
3. **spatial_audio_update** — recalculates attenuation and panning for every `SpatialAudioSink`.
4. **audio_cleanup** — detects removed `AudioPlayer` components or finished playback, drops sinks, and despawns entities in `DESPAWN` mode.

## 5. Open Questions
- Should streaming playback (decode-on-the-fly) be a separate asset type or a flag on `AudioSource`?
- What distance attenuation model should be default — inverse distance, linear, or configurable per-source?
- How should audio ducking (lowering music when dialogue plays) be expressed?

## Document History
| Version | Date | Description |
| :--- | :--- | :--- |
| 0.1.0 | 2026-03-25 | Initial draft from architecture analysis |
