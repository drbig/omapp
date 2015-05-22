require 'haml'

def partial(name, locals = {})
  source = File.read(File.join('frontend', 'src', 'partials', name.to_s + '.haml'))
  engine = Haml::Engine.new(source)
  engine.render(binding, locals: locals)
end

def js(name)
  File.read(File.join('frontend', 'build', name.to_s + '.min.js'))
end

def vars
  if ENV['OMA_BUILD'] == 'prod'
    path = File.join('frontend', 'build', 'vars_prod.min.js')
  else
    path = File.join('frontend', 'build', 'vars_devel.min.js')
  end
  File.read(path)
end

def css(name)
  File.read(File.join('frontend', 'build', name.to_s + '.css'))
end
